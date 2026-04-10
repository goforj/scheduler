package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
)

type stubLocker struct{}

func (stubLocker) Lock(ctx context.Context, key string) (gocron.Lock, error) {
	return stubLock{}, nil
}

type stubLock struct{}

func (stubLock) Unlock(ctx context.Context) error { return nil }

type testFoo struct{}

func (testFoo) sample(context.Context) error     { return nil }
func (*testFoo) ptrSample(context.Context) error { return nil }

func newTestScheduler(clock *clockwork.FakeClock) gocron.Scheduler {
	s, err := gocron.NewScheduler(gocron.WithClock(clock))
	if err != nil {
		panic(err)
	}
	s.Start()
	return s
}

func waitForJob(clock *clockwork.FakeClock, d time.Duration) {
	clock.Advance(d)
	time.Sleep(10 * time.Millisecond)
}

func sampleHandler(context.Context) error { return nil }

func ptrDur(d time.Duration) *time.Duration { return &d }

func TestEverySecond_Do(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	var called atomic.Bool

	newJobBuilder(s).
		EverySecond().
		Do(func(context.Context) error { called.Store(true); return nil })

	waitForJob(clock, time.Second)
	require.True(t, called.Load())
	s.Shutdown()
}

func TestWhenSkipAndEnvironments(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	var called atomic.Bool

	newJobBuilder(s).
		Environments("production").
		EverySecond().
		Do(func(context.Context) error { called.Store(true); return nil })
	require.Len(t, newJobBuilder(s).JobMetadata(), 0)

	newJobBuilder(s).
		EverySecond().
		When(func() bool { return false }).
		Do(func(context.Context) error { called.Store(true); return nil })
	require.False(t, called.Load())

	newJobBuilder(s).
		EverySecond().
		Skip(func() bool { return true }).
		Do(func(context.Context) error { called.Store(true); return nil })
	require.False(t, called.Load())
}

func TestScheduleBuilders_AllVariants(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	var oneSec = 1 * time.Second

	cases := []struct {
		name    string
		build   func() *JobBuilder
		wantErr bool
		wantDur *time.Duration
		wantExp string
	}{
		{"EverySecond", func() *JobBuilder { return newJobBuilder(nil).EverySecond() }, false, &oneSec, ""},
		{"EveryTenSecondsFluent", func() *JobBuilder { return newJobBuilder(nil).Every(10).Seconds() }, false, ptrDur(10 * time.Second), ""},
		{"DailyAt", func() *JobBuilder { return newJobBuilder(nil).DailyAt("13:00") }, false, nil, "0 13 * * *"},
		{"DailyAtInvalid", func() *JobBuilder { return newJobBuilder(nil).DailyAt("bad") }, true, nil, ""},
		{"TwiceDailyAt", func() *JobBuilder { return newJobBuilder(nil).TwiceDailyAt(1, 13, 15) }, false, nil, "15 1,13 * * *"},
		{"WeeklyOn", func() *JobBuilder { return newJobBuilder(nil).WeeklyOn(1, "8:00") }, false, nil, "0 8 * * 1"},
		{"WeeklyOnInvalid", func() *JobBuilder { return newJobBuilder(nil).WeeklyOn(1, "bad") }, true, nil, ""},
		{"MonthlyOn", func() *JobBuilder { return newJobBuilder(nil).MonthlyOn(5, "9:30") }, false, nil, "30 9 5 * *"},
		{"MonthlyOnInvalid", func() *JobBuilder { return newJobBuilder(nil).MonthlyOn(5, "x") }, true, nil, ""},
		{"TwiceMonthly", func() *JobBuilder { return newJobBuilder(nil).TwiceMonthly(1, 15, "10:00") }, false, nil, "0 10 1,15 * *"},
		{"TwiceMonthlyInvalid", func() *JobBuilder { return newJobBuilder(nil).TwiceMonthly(1, 15, "x") }, true, nil, ""},
		{"LastDayOfMonth", func() *JobBuilder { return newJobBuilder(nil).LastDayOfMonth("23:15") }, false, nil, "15 23 L * *"},
		{"LastDayOfMonthInvalid", func() *JobBuilder { return newJobBuilder(nil).LastDayOfMonth("x") }, true, nil, ""},
		{"QuarterlyOn", func() *JobBuilder { return newJobBuilder(nil).QuarterlyOn(3, "12:00") }, false, nil, "0 12 3 1,4,7,10 *"},
		{"QuarterlyOnInvalid", func() *JobBuilder { return newJobBuilder(nil).QuarterlyOn(3, "x") }, true, nil, ""},
		{"YearlyOn", func() *JobBuilder { return newJobBuilder(nil).YearlyOn(12, 25, "6:45") }, false, nil, "45 6 25 12 *"},
		{"YearlyOnInvalid", func() *JobBuilder { return newJobBuilder(nil).YearlyOn(12, 25, "x") }, true, nil, ""},
		{"HourlyAt", func() *JobBuilder { return newJobBuilder(nil).HourlyAt(5) }, false, nil, "5 * * * *"},
		{"Daily", func() *JobBuilder { return newJobBuilder(nil).Daily() }, false, nil, "0 0 * * *"},
		{"Weekly", func() *JobBuilder { return newJobBuilder(nil).Weekly() }, false, nil, "0 0 * * 0"},
		{"Quarterly", func() *JobBuilder { return newJobBuilder(nil).Quarterly() }, false, nil, "0 0 1 1,4,7,10 *"},
		{"Yearly", func() *JobBuilder { return newJobBuilder(nil).Yearly() }, false, nil, "0 0 1 1 *"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := tc.build()
			if tc.wantErr {
				require.Error(t, f.Error())
			} else {
				require.NoError(t, f.Error())
			}
			if tc.wantDur != nil {
				require.NotNil(t, f.duration)
				require.Equal(t, *tc.wantDur, *f.duration)
			}
			if tc.wantExp != "" {
				require.Equal(t, tc.wantExp, f.CronExpr())
			}
		})
	}
}

func TestCommandMetadataAndRunNowSkipsExec(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("SCHEDULER_TEST_NO_EXEC", "1")
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	f := newJobBuilder(s)
	f.Cron("0 0 * * *").Command("hello:world", "foo", "bar")

	job := f.Job()
	require.NotNil(t, job)
	require.NoError(t, job.RunNow())

	meta := f.JobMetadata()
	require.Len(t, meta, 1)
	for _, v := range meta {
		require.Equal(t, string(jobTargetCommand), v.TargetKind)
		require.Equal(t, "hello:world foo bar", v.Command)
		require.Equal(t, string(jobScheduleCron), v.ScheduleType)
		require.Equal(t, "0 0 * * *", v.Schedule)
		require.Contains(t, v.Tags, "env=test")
		require.Contains(t, v.Tags, "cron=0 0 * * *")
		require.Contains(t, v.Tags, "args=\"foo bar\"")
	}
}

func TestCommandNoArgsMetadataAndReset(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("SCHEDULER_TEST_NO_EXEC", "1")
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	f := newJobBuilder(s)
	f.Cron("0 0 * * *").Command("hello:world")

	require.NoError(t, f.Job().RunNow())

	meta := f.JobMetadata()
	require.Len(t, meta, 1)
	for _, v := range meta {
		require.Equal(t, "hello:world", v.Command)
		require.Equal(t, string(jobScheduleCron), v.ScheduleType)
		require.Equal(t, "0 0 * * *", v.Schedule)
		require.Contains(t, v.Tags, "env=test")
	}

	require.Equal(t, "", f.CronExpr())
	require.Nil(t, f.duration)
}

func TestExecMetadataAndRunner(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	var gotExe string
	var gotArgs []string
	done := make(chan struct{})
	runner := CommandRunnerFunc(func(_ context.Context, exe string, args []string) error {
		gotExe = exe
		gotArgs = append([]string{}, args...)
		close(done)
		return nil
	})

	f := newJobBuilder(s).WithCommandRunner(runner)
	f.Cron("0 0 * * *").Exec("/usr/bin/env", "echo", "hello")

	require.NotNil(t, f.Job())
	require.NoError(t, f.Job().RunNow())
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("exec runner did not run")
	}
	require.Equal(t, "/usr/bin/env", gotExe)
	require.Equal(t, []string{"echo", "hello"}, gotArgs)

	meta := f.JobMetadata()
	require.Len(t, meta, 1)
	for _, v := range meta {
		require.Equal(t, string(jobTargetExec), v.TargetKind)
		require.Equal(t, "/usr/bin/env echo hello", v.Command)
		require.Equal(t, string(jobScheduleCron), v.ScheduleType)
		require.Equal(t, "0 0 * * *", v.Schedule)
		require.Contains(t, v.Tags, "env=test")
		require.Contains(t, v.Tags, "args=\"echo hello\"")
	}
}

func TestFunctionMetadataFriendlyNameAndRetainState(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	var called atomic.Int32
	fn := func(context.Context) error {
		called.Add(1)
		return nil
	}

	f := newJobBuilder(s).
		RetainState().
		EverySecond().
		Do(fn)

	require.NotNil(t, f.Job())
	meta := f.JobMetadata()
	require.Len(t, meta, 1)
	for _, v := range meta {
		require.Equal(t, string(jobTargetFunction), v.TargetKind)
		require.Equal(t, "scheduler.TestFunctionMetadataFriendlyNameAndRetainState (anon func)", v.Handler)
	}

	f.Do(fn)
	require.Equal(t, int32(0), called.Load())
	require.NotNil(t, f.Job())
	require.Nil(t, f.duration)
}

func TestIntervalMetadataAndTags(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	f := newJobBuilder(s).
		EverySecond().
		Do(sampleHandler)

	meta := f.JobMetadata()
	require.Len(t, meta, 1)
	for _, v := range meta {
		require.Equal(t, string(jobScheduleInterval), v.ScheduleType)
		require.Equal(t, "1s", v.Schedule)
		require.Equal(t, string(jobTargetFunction), v.TargetKind)
		require.Equal(t, "scheduler.sampleHandler", v.Handler)
		require.Contains(t, v.Tags, "env=test")
		require.Contains(t, v.Tags, "interval=1s")
	}
}

func TestJobMetadataCopyIsDefensive(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	f := newJobBuilder(s).
		EverySecond().
		Do(sampleHandler)

	first := f.JobMetadata()
	for id := range first {
		first[id] = JobMetadata{Handler: "tampered"}
		delete(first, id)
	}

	second := f.JobMetadata()
	require.Len(t, second, 1)
	for _, v := range second {
		require.Equal(t, "scheduler.sampleHandler", v.Handler)
	}
}

func TestOverlapAndLockerOptions(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	f := newJobBuilder(s).
		WithoutOverlapping().
		WithoutOverlappingWithLocker(stubLocker{}).
		EverySecond().
		Do(func(context.Context) error { return nil })

	require.NotNil(t, f.Job())
}

func TestDescribeScheduleVariants(t *testing.T) {
	f := newJobBuilder(nil)
	kind, val := f.describeSchedule()
	require.Equal(t, jobScheduleUnknown, kind)
	require.Equal(t, "", val)

	f.cronExpr = "0 0 * * *"
	kind, val = f.describeSchedule()
	require.Equal(t, jobScheduleCron, kind)
	require.Equal(t, "0 0 * * *", val)

	d := time.Second
	f.duration = &d
	f.cronExpr = ""
	kind, val = f.describeSchedule()
	require.Equal(t, jobScheduleInterval, kind)
	require.Equal(t, "1s", val)
}

func TestFriendlyFuncNameVariants(t *testing.T) {
	require.Equal(t, "", friendlyFuncName(nil))
	require.Contains(t, friendlyFuncName(sampleHandler), "scheduler.sampleHandler")

	fn := func() {}
	require.Contains(t, friendlyFuncName(fn), "anon func")

	f := testFoo{}
	require.Contains(t, friendlyFuncName(f.sample), "testFoo.sample")
	require.Contains(t, friendlyFuncName((&testFoo{}).sample), "testFoo.sample")
	require.Contains(t, friendlyFuncName((&testFoo{}).ptrSample), "testFoo.ptrSample")
}

func TestBuildCommandString(t *testing.T) {
	require.Equal(t, "", buildCommandString("", nil))
	require.Equal(t, "cmd", buildCommandString("cmd", nil))
	require.Equal(t, "cmd arg1 arg2", buildCommandString("cmd", []string{"arg1", "arg2"}))
}

func TestTimezoneSetter(t *testing.T) {
	f := newJobBuilder(nil)
	out := f.Timezone("UTC")
	require.Equal(t, "UTC", out.timezone)
}

func TestParseHourMinuteValidation(t *testing.T) {
	_, _, err := parseHourMinute("bad")
	require.Error(t, err)

	_, _, err = parseHourMinute("aa:bb")
	require.Error(t, err)

	_, _, err = parseHourMinute("12:xx")
	require.Error(t, err)

	h, m, err := parseHourMinute("13:45")
	require.NoError(t, err)
	require.Equal(t, 13, h)
	require.Equal(t, 45, m)
}

func TestAllDurationBuilders(t *testing.T) {
	tests := []struct {
		name string
		fn   func() *JobBuilder
		want time.Duration
	}{
		{"Minutes", func() *JobBuilder { return newJobBuilder(nil).Every(3).Minutes() }, 3 * time.Minute},
		{"Hours", func() *JobBuilder { return newJobBuilder(nil).Every(2).Hours() }, 2 * time.Hour},
		{"EveryTwoSeconds", func() *JobBuilder { return newJobBuilder(nil).EveryTwoSeconds() }, 2 * time.Second},
		{"EveryFiveSeconds", func() *JobBuilder { return newJobBuilder(nil).EveryFiveSeconds() }, 5 * time.Second},
		{"EveryTenSeconds", func() *JobBuilder { return newJobBuilder(nil).EveryTenSeconds() }, 10 * time.Second},
		{"EveryFifteenSeconds", func() *JobBuilder { return newJobBuilder(nil).EveryFifteenSeconds() }, 15 * time.Second},
		{"EveryTwentySeconds", func() *JobBuilder { return newJobBuilder(nil).EveryTwentySeconds() }, 20 * time.Second},
		{"EveryThirtySeconds", func() *JobBuilder { return newJobBuilder(nil).EveryThirtySeconds() }, 30 * time.Second},
		{"EveryMinute", func() *JobBuilder { return newJobBuilder(nil).EveryMinute() }, time.Minute},
		{"EveryTwoMinutes", func() *JobBuilder { return newJobBuilder(nil).EveryTwoMinutes() }, 2 * time.Minute},
		{"EveryThreeMinutes", func() *JobBuilder { return newJobBuilder(nil).EveryThreeMinutes() }, 3 * time.Minute},
		{"EveryFourMinutes", func() *JobBuilder { return newJobBuilder(nil).EveryFourMinutes() }, 4 * time.Minute},
		{"EveryFiveMinutes", func() *JobBuilder { return newJobBuilder(nil).EveryFiveMinutes() }, 5 * time.Minute},
		{"EveryTenMinutes", func() *JobBuilder { return newJobBuilder(nil).EveryTenMinutes() }, 10 * time.Minute},
		{"EveryFifteenMinutes", func() *JobBuilder { return newJobBuilder(nil).EveryFifteenMinutes() }, 15 * time.Minute},
		{"EveryThirtyMinutes", func() *JobBuilder { return newJobBuilder(nil).EveryThirtyMinutes() }, 30 * time.Minute},
		{"Hourly", func() *JobBuilder { return newJobBuilder(nil).Hourly() }, time.Hour},
		{"EveryOddHour", func() *JobBuilder { return newJobBuilder(nil).EveryOddHour(5) }, 0},
		{"EveryTwoHours", func() *JobBuilder { return newJobBuilder(nil).EveryTwoHours(10) }, 0},
		{"EveryThreeHours", func() *JobBuilder { return newJobBuilder(nil).EveryThreeHours(7) }, 0},
		{"EveryFourHours", func() *JobBuilder { return newJobBuilder(nil).EveryFourHours(9) }, 0},
		{"EverySixHours", func() *JobBuilder { return newJobBuilder(nil).EverySixHours(11) }, 0},
		{"TwiceDaily", func() *JobBuilder { return newJobBuilder(nil).TwiceDaily(1, 13) }, 0},
		{"Weekly", func() *JobBuilder { return newJobBuilder(nil).Weekly() }, 0},
		{"Monthly", func() *JobBuilder { return newJobBuilder(nil).Monthly() }, 0},
		{"Quarterly", func() *JobBuilder { return newJobBuilder(nil).Quarterly() }, 0},
		{"Yearly", func() *JobBuilder { return newJobBuilder(nil).Yearly() }, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := tc.fn()
			if tc.want > 0 {
				require.NotNil(t, f.duration)
				require.Equal(t, tc.want, *f.duration)
			} else {
				require.NotEmpty(t, f.CronExpr())
			}
		})
	}
}

func TestCommandWithExistingErrorSkipsScheduling(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	f := newJobBuilder(s)
	f.err = fmt.Errorf("preset error")
	f.Command("noop")
	require.Nil(t, f.Job())
	require.Len(t, f.JobMetadata(), 0)
}

func TestWeeklyOnHourAndMinuteErrors(t *testing.T) {
	f := newJobBuilder(nil).WeeklyOn(1, "bad:30")
	require.Error(t, f.Error())

	f2 := newJobBuilder(nil).WeeklyOn(1, "10:xx")
	require.Error(t, f2.Error())
}

func TestDoWithoutScheduleErrors(t *testing.T) {
	s := newTestScheduler(clockwork.NewFakeClock())
	defer s.Shutdown()

	f := newJobBuilder(s).Do(func(context.Context) error { return nil })
	require.Error(t, f.Error())
	require.Nil(t, f.Job())
}

type stubJob struct {
	id   uuid.UUID
	name string
	tags []string
}

func (j stubJob) ID() uuid.UUID                     { return j.id }
func (j stubJob) LastRun() (time.Time, error)       { return time.Time{}, nil }
func (j stubJob) Name() string                      { return j.name }
func (j stubJob) NextRun() (time.Time, error)       { return time.Now(), nil }
func (j stubJob) NextRuns(int) ([]time.Time, error) { return nil, nil }
func (j stubJob) RunNow() error                     { return nil }
func (j stubJob) Tags() []string                    { return j.tags }

func TestRecordJobFallbacks(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	f := newJobBuilder(nil)
	f.targetKind = jobTargetCommand
	f.name = "cmd"
	f.commandArgs = []string{"a"}
	d := time.Second
	f.duration = &d

	j := stubJob{id: uuid.New()}
	f.recordJob(j, sampleHandler)

	meta := f.JobMetadata()
	require.Len(t, meta, 1)
	for _, v := range meta {
		require.Equal(t, "cmd", v.Name)
		require.Equal(t, "cmd a", v.Command)
		require.Equal(t, string(jobScheduleInterval), v.ScheduleType)
		require.Equal(t, "1s", v.Schedule)
	}
}

func TestCommandRunNowWithLateError(t *testing.T) {
	t.Setenv("SCHEDULER_TEST_NO_EXEC", "1")
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	f := newJobBuilder(s)
	f.Cron("0 0 * * *").Command("hello:world")

	f.err = fmt.Errorf("late err")
	require.NoError(t, f.Job().RunNow())
}

func TestDoWithPresetErrorSkipsScheduling(t *testing.T) {
	s := newTestScheduler(clockwork.NewFakeClock())
	defer s.Shutdown()

	f := newJobBuilder(s)
	f.err = fmt.Errorf("boom")
	f.Do(func(context.Context) error { return nil })
	require.Nil(t, f.Job())
}

func TestHooksRunForFunctionTasks(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	var before, after, success atomic.Bool

	f := newJobBuilder(s)
	f.Before(func(context.Context) { before.Store(true) }).
		After(func(context.Context) { after.Store(true) }).
		OnSuccess(func(context.Context) { success.Store(true) }).
		EverySecond().
		Do(func(context.Context) error { return nil })

	waitForJob(clock, 3*time.Second)

	require.True(t, before.Load())
	require.True(t, after.Load())
	require.True(t, success.Load())
}

func TestFailureHookReceivesError(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	var failureCalled atomic.Bool
	var receivedErr atomic.Pointer[error]

	expectedErr := context.DeadlineExceeded

	newJobBuilder(s).
		OnFailure(func(_ context.Context, err error) {
			failureCalled.Store(true)
			receivedErr.Store(&err)
		}).
		EverySecond().
		Do(func(context.Context) error { return expectedErr })

	waitForJob(clock, 3*time.Second)

	require.True(t, failureCalled.Load())
	got := receivedErr.Load()
	require.NotNil(t, got)
	require.ErrorIs(t, *got, expectedErr)
}

func TestDaysOfMonth(t *testing.T) {
	f := newJobBuilder(nil).DaysOfMonth([]int{1, 10, 20}, "13:00")
	require.Equal(t, "0 13 1,10,20 * *", f.CronExpr())
}

func TestDayConstraintsAndBetween(t *testing.T) {
	orig := nowFunc
	defer func() { nowFunc = orig }()

	nowFunc = func() time.Time {
		return time.Date(2024, 9, 16, 9, 0, 0, 0, time.UTC) // Monday 09:00 UTC
	}

	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	var weekdayCalled, weekendCalled, betweenCalled, unlessBetweenSkipped atomic.Bool

	newJobBuilder(s).
		Timezone("UTC").
		Weekdays().
		EverySecond().
		Do(func(context.Context) error { weekdayCalled.Store(true); return nil })

	newJobBuilder(s).
		Timezone("UTC").
		Weekends().
		EverySecond().
		Do(func(context.Context) error { weekendCalled.Store(true); return nil })

	newJobBuilder(s).
		Timezone("UTC").
		Between("08:00", "10:00").
		EverySecond().
		Do(func(context.Context) error { betweenCalled.Store(true); return nil })

	newJobBuilder(s).
		Timezone("UTC").
		UnlessBetween("08:00", "10:00").
		EverySecond().
		Do(func(context.Context) error { unlessBetweenSkipped.Store(true); return nil })

	waitForJob(clock, time.Second)

	require.True(t, weekdayCalled.Load())
	require.False(t, weekendCalled.Load())
	require.True(t, betweenCalled.Load())
	require.False(t, unlessBetweenSkipped.Load())
}

func TestSpecificDayHelpers(t *testing.T) {
	orig := nowFunc
	defer func() { nowFunc = orig }()
	nowFunc = func() time.Time {
		return time.Date(2024, 9, 17, 12, 0, 0, 0, time.UTC) // Tuesday
	}

	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	var mondayCalled, tuesdayCalled atomic.Bool

	newJobBuilder(s).
		Timezone("UTC").
		Mondays().
		EverySecond().
		Do(func(context.Context) error { mondayCalled.Store(true); return nil })

	newJobBuilder(s).
		Timezone("UTC").
		Tuesdays().
		EverySecond().
		Do(func(context.Context) error { tuesdayCalled.Store(true); return nil })

	waitForJob(clock, time.Second)

	require.False(t, mondayCalled.Load())
	require.True(t, tuesdayCalled.Load())
}

func TestSundayHelper(t *testing.T) {
	orig := nowFunc
	defer func() { nowFunc = orig }()
	nowFunc = func() time.Time {
		return time.Date(2024, 9, 15, 9, 0, 0, 0, time.UTC) // Sunday
	}

	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	var sundayCalled, saturdayCalled atomic.Bool

	newJobBuilder(s).Timezone("UTC").Sundays().EverySecond().Do(func(context.Context) error { sundayCalled.Store(true); return nil })
	newJobBuilder(s).Timezone("UTC").Saturdays().EverySecond().Do(func(context.Context) error { saturdayCalled.Store(true); return nil })

	waitForJob(clock, time.Second)

	require.True(t, sundayCalled.Load())
	require.False(t, saturdayCalled.Load())
}

func TestMidweekHelpers(t *testing.T) {
	orig := nowFunc
	defer func() { nowFunc = orig }()
	nowFunc = func() time.Time {
		return time.Date(2024, 9, 19, 12, 0, 0, 0, time.UTC) // Thursday
	}

	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	var wedCalled, thuCalled, friCalled atomic.Bool

	newJobBuilder(s).Timezone("UTC").Wednesdays().EverySecond().Do(func(context.Context) error { wedCalled.Store(true); return nil })
	newJobBuilder(s).Timezone("UTC").Thursdays().EverySecond().Do(func(context.Context) error { thuCalled.Store(true); return nil })
	newJobBuilder(s).Timezone("UTC").Fridays().EverySecond().Do(func(context.Context) error { friCalled.Store(true); return nil })

	waitForJob(clock, time.Second)

	require.False(t, wedCalled.Load())
	require.True(t, thuCalled.Load())
	require.False(t, friCalled.Load())
}

func TestTimezoneAppliedToCron(t *testing.T) {
	t.Setenv("SCHEDULER_TEST_NO_EXEC", "1")
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	f := newJobBuilder(s).
		Timezone("America/Chicago").
		Cron("0 12 * * *").
		Command("hello:world")

	meta := f.JobMetadata()
	require.Len(t, meta, 1)
	for _, v := range meta {
		require.Equal(t, string(jobScheduleCron), v.ScheduleType)
		require.Contains(t, v.Schedule, "CRON_TZ=America/Chicago")
	}
}

func TestDaysOfMonthInvalid(t *testing.T) {
	f := newJobBuilder(nil).DaysOfMonth([]int{1, 2}, "bad")
	require.Error(t, f.Error())
}

func TestBetweenCrossesMidnight(t *testing.T) {
	orig := nowFunc
	defer func() { nowFunc = orig }()
	nowFunc = func() time.Time {
		return time.Date(2024, 9, 17, 1, 0, 0, 0, time.UTC) // 01:00 UTC
	}

	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	var betweenCalled, unlessBetweenCalled atomic.Bool

	newJobBuilder(s).Timezone("UTC").Between("23:00", "02:00").EverySecond().Do(func(context.Context) error { betweenCalled.Store(true); return nil })
	newJobBuilder(s).Timezone("UTC").UnlessBetween("23:00", "02:00").EverySecond().Do(func(context.Context) error { unlessBetweenCalled.Store(true); return nil })

	waitForJob(clock, time.Second)

	require.True(t, betweenCalled.Load())
	require.False(t, unlessBetweenCalled.Load())
}

func TestAddWhenSkipComposition(t *testing.T) {
	f := newJobBuilder(nil)
	f.When(func() bool { return true })
	f.When(func() bool { return false })
	require.False(t, f.whenFunc())

	f.Skip(func() bool { return false })
	f.Skip(func() bool { return true })
	require.True(t, f.skipFunc())
}

func TestCommandRunnerFunc_Run(t *testing.T) {
	var called bool
	runner := CommandRunnerFunc(func(_ context.Context, exe string, args []string) error {
		called = true
		require.Equal(t, "bin", exe)
		require.Equal(t, []string{"a", "b"}, args)
		return nil
	})

	require.NoError(t, runner.Run(context.Background(), "bin", []string{"a", "b"}))
	require.True(t, called)
}

func TestLockerFuncAndLockFunc(t *testing.T) {
	var locked bool

	locker := LockerFunc(func(ctx context.Context, key string) (gocron.Lock, error) {
		locked = true
		require.Equal(t, "job", key)
		return LockFunc(func(context.Context) error { return nil }), nil
	})

	lock, err := locker.Lock(context.Background(), "job")
	require.NoError(t, err)
	require.True(t, locked)
	require.NoError(t, lock.Unlock(context.Background()))
}

func TestBetweenInvalidInputs(t *testing.T) {
	b := newJobBuilder(nil)
	b.Between("bad", "10:00")
	require.Error(t, b.Error())

	b = newJobBuilder(nil)
	b.UnlessBetween("09:00", "bad")
	require.Error(t, b.Error())
}

func TestTimeInRangeCrossesMidnightEarly(t *testing.T) {
	loc := time.UTC
	now := time.Date(2024, 1, 2, 23, 0, 0, 0, loc) // after start window
	require.True(t, timeInRange(now, 22, 0, 6, 0))
}

func TestRenderHelpersEdges(t *testing.T) {
	require.Equal(t, jobStyleMissingInfo.Render("unknown"), renderJobType(""))
	require.Equal(t, jobStyleSchedule.Render("every 1s"), renderSchedule("1s", jobScheduleInterval))
	require.Equal(t, jobStyleMissingInfo.Render("-"), renderTarget("", jobTargetCommand, "name"))
	require.Equal(t, jobStyleMissingInfo.Render("-"), renderTags([]string{"cron=*"}))

	sched, kind := scheduleFromTags([]string{"foo=bar"})
	require.Equal(t, "", sched)
	require.Equal(t, jobScheduleUnknown, kind)

	entry := &jobEntry{Name: "", Target: "", NextRun: time.Time{}}
	require.Equal(t, "(unnamed)", nameForEntry(entry))
	require.Equal(t, 1, maxRelativeWidth([]*jobEntry{entry}))
}

func TestAddWhenSkipNilAndCombine(t *testing.T) {
	b := newJobBuilder(nil)
	b.addWhen(nil)
	require.Nil(t, b.whenFunc)

	first := false
	b.addWhen(func() bool { first = true; return true })
	require.NotNil(t, b.whenFunc)

	second := false
	b.addWhen(func() bool { second = true; return false })
	require.False(t, b.whenFunc())
	require.True(t, first)
	require.True(t, second)

	b.addSkip(nil)
	require.Nil(t, b.skipFunc)

	b.addSkip(func() bool { return true })
	require.True(t, b.skipFunc())

	b.addSkip(func() bool { return false })
	require.True(t, b.skipFunc())
}

func TestLocationInvalidZoneFallsBack(t *testing.T) {
	b := newJobBuilder(nil).Timezone("Invalid/Zone")
	require.Equal(t, time.Local, b.location())
}

func TestFriendlyFuncNameNil(t *testing.T) {
	var f func()
	require.Equal(t, "", friendlyFuncName(f))
}

func TestBetweenUnlessBetweenValid(t *testing.T) {
	fixed := func() time.Time { return time.Date(2024, 1, 2, 1, 30, 0, 0, time.UTC) }

	b := newJobBuilder(nil).Timezone("UTC").WithNowFunc(fixed).Between("01:00", "02:00")
	require.NoError(t, b.Error())
	require.NotNil(t, b.whenFunc)
	// 01:30 is inside window
	require.True(t, b.whenFunc())

	b = newJobBuilder(nil).Timezone("UTC").WithNowFunc(fixed).UnlessBetween("22:00", "23:00")
	require.NoError(t, b.Error())
	require.NotNil(t, b.skipFunc)
	// 01:30 is outside skip window
	require.False(t, b.skipFunc())
}

func TestCommandRunInBackgroundBranch(t *testing.T) {
	t.Setenv("SCHEDULER_TEST_NO_EXEC", "1")
	b := newJobBuilder(newTestScheduler(clockwork.NewFakeClock()))
	b.RunInBackground().Cron("0 0 * * *").Command("noop")
	require.NotNil(t, b.Job())
}

func TestDoWithFiltersSkipping(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	s := newTestScheduler(clockwork.NewFakeClock())
	defer s.Shutdown()

	b := newJobBuilder(s).Environments("production").EverySecond().Do(func(context.Context) error { return nil })
	require.Nil(t, b.Job())

	b = newJobBuilder(s).EverySecond().When(func() bool { return false }).Do(func(context.Context) error { return nil })
	require.Nil(t, b.Job())
}

func TestDoWithoutSchedule(t *testing.T) {
	b := newJobBuilder(nil)
	b.Do(func(context.Context) error { return nil })
	require.Error(t, b.Error())
}

func TestRecordJobDefaults(t *testing.T) {
	b := newJobBuilder(nil)
	b.targetKind = ""
	b.state = nil

	j := stubJob{id: uuid.New()}
	b.recordJob(j, sampleHandler)

	meta := b.JobMetadata()
	require.Len(t, meta, 1)
	for _, v := range meta {
		require.Equal(t, string(jobTargetFunction), v.TargetKind)
		require.Equal(t, "scheduler.sampleHandler", v.Handler)
	}
}

type stubSchedulerAdapter struct {
	jobs []gocron.Job
}

func (s stubSchedulerAdapter) NewJob(job gocron.JobDefinition, task gocron.Task, options ...gocron.JobOption) (gocron.Job, error) {
	return nil, errors.New("not implemented")
}
func (s stubSchedulerAdapter) Start()             {}
func (s stubSchedulerAdapter) Shutdown() error    { return nil }
func (s stubSchedulerAdapter) Jobs() []gocron.Job { return s.jobs }

func TestGetJobsListCoverage(t *testing.T) {
	var nilBuilder *JobBuilder
	require.Nil(t, nilBuilder.getJobsList())

	cronJob := stubJob{
		id:   uuid.New(),
		name: "",
		tags: []string{"cron=0 0 * * *"},
	}
	b := newJobBuilder(nil)
	b.scheduler = stubSchedulerAdapter{jobs: []gocron.Job{cronJob}}
	entries := b.getJobsList()
	require.Len(t, entries, 1)
	require.Equal(t, jobScheduleCron, entries[0].ScheduleType)
	require.Equal(t, "(unnamed)", entries[0].Name)
}

func TestRenderHelpersMoreBranches(t *testing.T) {
	require.Equal(t, jobStyleSchedule.Render("cron 0 0 * * *"), renderSchedule("0 0 * * *", jobScheduleCron))
	require.Equal(t, jobStyleSchedule.Render("foo"), renderSchedule("foo", jobScheduleUnknown))
	require.Equal(t, jobStyleMissingInfo.Render("-"), renderTarget("name", jobTargetCommand, "name"))
	require.Equal(t, jobStyleMissingInfo.Render("-"), renderSchedule("", jobScheduleUnknown))

	require.Contains(t, renderNextRunWithWidth(time.Time{}, errors.New("boom"), 5), "boom")

	entry := &jobEntry{
		Name:    "a",
		NextRun: time.Now().Add(2 * time.Hour),
	}
	require.Greater(t, maxRelativeWidth([]*jobEntry{nil, entry}), 0)

	errEntry := &jobEntry{NextRunErr: errors.New("fail")}
	require.Equal(t, 1, maxRelativeWidth([]*jobEntry{errEntry}))
}

func TestFriendlyFuncNameAnonAndLocationValid(t *testing.T) {
	name := friendlyFuncName(func() {})
	require.Contains(t, name, "anon func")

	loc := newJobBuilder(nil).Timezone("America/New_York").location()
	require.Equal(t, "America/New_York", loc.String())
}
