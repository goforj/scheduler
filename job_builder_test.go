package scheduler

import (
	"context"
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

func (testFoo) sample()     {}
func (*testFoo) ptrSample() {}

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

func sampleHandler() {}

func ptrDur(d time.Duration) *time.Duration { return &d }

func TestEverySecond_Do(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	var called atomic.Bool

	NewJobBuilder(s).
		EverySecond().
		Do(func() { called.Store(true) })

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

	NewJobBuilder(s).
		Environments("production").
		EverySecond().
		Do(func() { called.Store(true) })
	require.Len(t, NewJobBuilder(s).JobMetadata(), 0)

	NewJobBuilder(s).
		EverySecond().
		When(func() bool { return false }).
		Do(func() { called.Store(true) })
	require.False(t, called.Load())

	NewJobBuilder(s).
		EverySecond().
		Skip(func() bool { return true }).
		Do(func() { called.Store(true) })
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
		{"EverySecond", func() *JobBuilder { return NewJobBuilder(nil).EverySecond() }, false, &oneSec, ""},
		{"EveryTenSecondsFluent", func() *JobBuilder { return NewJobBuilder(nil).Every(10).Seconds() }, false, ptrDur(10 * time.Second), ""},
		{"DailyAt", func() *JobBuilder { return NewJobBuilder(nil).DailyAt("13:00") }, false, nil, "0 13 * * *"},
		{"DailyAtInvalid", func() *JobBuilder { return NewJobBuilder(nil).DailyAt("bad") }, true, nil, ""},
		{"TwiceDailyAt", func() *JobBuilder { return NewJobBuilder(nil).TwiceDailyAt(1, 13, 15) }, false, nil, "15 1,13 * * *"},
		{"WeeklyOn", func() *JobBuilder { return NewJobBuilder(nil).WeeklyOn(1, "8:00") }, false, nil, "0 8 * * 1"},
		{"WeeklyOnInvalid", func() *JobBuilder { return NewJobBuilder(nil).WeeklyOn(1, "bad") }, true, nil, ""},
		{"MonthlyOn", func() *JobBuilder { return NewJobBuilder(nil).MonthlyOn(5, "9:30") }, false, nil, "30 9 5 * *"},
		{"MonthlyOnInvalid", func() *JobBuilder { return NewJobBuilder(nil).MonthlyOn(5, "x") }, true, nil, ""},
		{"TwiceMonthly", func() *JobBuilder { return NewJobBuilder(nil).TwiceMonthly(1, 15, "10:00") }, false, nil, "0 10 1,15 * *"},
		{"TwiceMonthlyInvalid", func() *JobBuilder { return NewJobBuilder(nil).TwiceMonthly(1, 15, "x") }, true, nil, ""},
		{"LastDayOfMonth", func() *JobBuilder { return NewJobBuilder(nil).LastDayOfMonth("23:15") }, false, nil, "15 23 L * *"},
		{"LastDayOfMonthInvalid", func() *JobBuilder { return NewJobBuilder(nil).LastDayOfMonth("x") }, true, nil, ""},
		{"QuarterlyOn", func() *JobBuilder { return NewJobBuilder(nil).QuarterlyOn(3, "12:00") }, false, nil, "0 12 3 1,4,7,10 *"},
		{"QuarterlyOnInvalid", func() *JobBuilder { return NewJobBuilder(nil).QuarterlyOn(3, "x") }, true, nil, ""},
		{"YearlyOn", func() *JobBuilder { return NewJobBuilder(nil).YearlyOn(12, 25, "6:45") }, false, nil, "45 6 25 12 *"},
		{"YearlyOnInvalid", func() *JobBuilder { return NewJobBuilder(nil).YearlyOn(12, 25, "x") }, true, nil, ""},
		{"HourlyAt", func() *JobBuilder { return NewJobBuilder(nil).HourlyAt(5) }, false, nil, "5 * * * *"},
		{"Daily", func() *JobBuilder { return NewJobBuilder(nil).Daily() }, false, nil, "0 0 * * *"},
		{"Weekly", func() *JobBuilder { return NewJobBuilder(nil).Weekly() }, false, nil, "0 0 * * 0"},
		{"Quarterly", func() *JobBuilder { return NewJobBuilder(nil).Quarterly() }, false, nil, "0 0 1 1,4,7,10 *"},
		{"Yearly", func() *JobBuilder { return NewJobBuilder(nil).Yearly() }, false, nil, "0 0 1 1 *"},
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

	f := NewJobBuilder(s)
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

	f := NewJobBuilder(s)
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

func TestFunctionMetadataFriendlyNameAndRetainState(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	var called atomic.Int32
	fn := func() { called.Add(1) }

	f := NewJobBuilder(s).
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

	f := NewJobBuilder(s).
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

	f := NewJobBuilder(s).
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

	f := NewJobBuilder(s).
		WithoutOverlapping().
		WithoutOverlappingWithLocker(stubLocker{}).
		EverySecond().
		Do(func() {})

	require.NotNil(t, f.Job())
}

func TestDescribeScheduleVariants(t *testing.T) {
	f := NewJobBuilder(nil)
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
	require.Contains(t, friendlyFuncName(f.sample), "testFoo")
	require.NotEmpty(t, friendlyFuncName((&testFoo{}).sample))
	require.Contains(t, friendlyFuncName((&testFoo{}).ptrSample), "testFoo")
}

func TestBuildCommandString(t *testing.T) {
	require.Equal(t, "", buildCommandString("", nil))
	require.Equal(t, "cmd", buildCommandString("cmd", nil))
	require.Equal(t, "cmd arg1 arg2", buildCommandString("cmd", []string{"arg1", "arg2"}))
}

func TestTimezoneSetter(t *testing.T) {
	f := NewJobBuilder(nil)
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
		{"Minutes", func() *JobBuilder { return NewJobBuilder(nil).Every(3).Minutes() }, 3 * time.Minute},
		{"Hours", func() *JobBuilder { return NewJobBuilder(nil).Every(2).Hours() }, 2 * time.Hour},
		{"EveryTwoSeconds", func() *JobBuilder { return NewJobBuilder(nil).EveryTwoSeconds() }, 2 * time.Second},
		{"EveryFiveSeconds", func() *JobBuilder { return NewJobBuilder(nil).EveryFiveSeconds() }, 5 * time.Second},
		{"EveryTenSeconds", func() *JobBuilder { return NewJobBuilder(nil).EveryTenSeconds() }, 10 * time.Second},
		{"EveryFifteenSeconds", func() *JobBuilder { return NewJobBuilder(nil).EveryFifteenSeconds() }, 15 * time.Second},
		{"EveryTwentySeconds", func() *JobBuilder { return NewJobBuilder(nil).EveryTwentySeconds() }, 20 * time.Second},
		{"EveryThirtySeconds", func() *JobBuilder { return NewJobBuilder(nil).EveryThirtySeconds() }, 30 * time.Second},
		{"EveryMinute", func() *JobBuilder { return NewJobBuilder(nil).EveryMinute() }, time.Minute},
		{"EveryTwoMinutes", func() *JobBuilder { return NewJobBuilder(nil).EveryTwoMinutes() }, 2 * time.Minute},
		{"EveryThreeMinutes", func() *JobBuilder { return NewJobBuilder(nil).EveryThreeMinutes() }, 3 * time.Minute},
		{"EveryFourMinutes", func() *JobBuilder { return NewJobBuilder(nil).EveryFourMinutes() }, 4 * time.Minute},
		{"EveryFiveMinutes", func() *JobBuilder { return NewJobBuilder(nil).EveryFiveMinutes() }, 5 * time.Minute},
		{"EveryTenMinutes", func() *JobBuilder { return NewJobBuilder(nil).EveryTenMinutes() }, 10 * time.Minute},
		{"EveryFifteenMinutes", func() *JobBuilder { return NewJobBuilder(nil).EveryFifteenMinutes() }, 15 * time.Minute},
		{"EveryThirtyMinutes", func() *JobBuilder { return NewJobBuilder(nil).EveryThirtyMinutes() }, 30 * time.Minute},
		{"Hourly", func() *JobBuilder { return NewJobBuilder(nil).Hourly() }, time.Hour},
		{"EveryOddHour", func() *JobBuilder { return NewJobBuilder(nil).EveryOddHour(5) }, 0},
		{"EveryTwoHours", func() *JobBuilder { return NewJobBuilder(nil).EveryTwoHours(10) }, 0},
		{"EveryThreeHours", func() *JobBuilder { return NewJobBuilder(nil).EveryThreeHours(7) }, 0},
		{"EveryFourHours", func() *JobBuilder { return NewJobBuilder(nil).EveryFourHours(9) }, 0},
		{"EverySixHours", func() *JobBuilder { return NewJobBuilder(nil).EverySixHours(11) }, 0},
		{"TwiceDaily", func() *JobBuilder { return NewJobBuilder(nil).TwiceDaily(1, 13) }, 0},
		{"Weekly", func() *JobBuilder { return NewJobBuilder(nil).Weekly() }, 0},
		{"Monthly", func() *JobBuilder { return NewJobBuilder(nil).Monthly() }, 0},
		{"Quarterly", func() *JobBuilder { return NewJobBuilder(nil).Quarterly() }, 0},
		{"Yearly", func() *JobBuilder { return NewJobBuilder(nil).Yearly() }, 0},
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

	f := NewJobBuilder(s)
	f.err = fmt.Errorf("preset error")
	f.Command("noop")
	require.Nil(t, f.Job())
	require.Len(t, f.JobMetadata(), 0)
}

func TestWeeklyOnHourAndMinuteErrors(t *testing.T) {
	f := NewJobBuilder(nil).WeeklyOn(1, "bad:30")
	require.Error(t, f.Error())

	f2 := NewJobBuilder(nil).WeeklyOn(1, "10:xx")
	require.Error(t, f2.Error())
}

func TestDoWithoutScheduleErrors(t *testing.T) {
	s := newTestScheduler(clockwork.NewFakeClock())
	defer s.Shutdown()

	f := NewJobBuilder(s).Do(func() {})
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
	f := NewJobBuilder(nil)
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

	f := NewJobBuilder(s)
	f.Cron("0 0 * * *").Command("hello:world")

	f.err = fmt.Errorf("late err")
	require.NoError(t, f.Job().RunNow())
}

func TestDoWithPresetErrorSkipsScheduling(t *testing.T) {
	s := newTestScheduler(clockwork.NewFakeClock())
	defer s.Shutdown()

	f := NewJobBuilder(s)
	f.err = fmt.Errorf("boom")
	f.Do(func() {})
	require.Nil(t, f.Job())
}

func TestHooksRunForFunctionTasks(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	var before, after, success atomic.Bool

	f := NewJobBuilder(s)
	f.Before(func() { before.Store(true) }).
		After(func() { after.Store(true) }).
		OnSuccess(func() { success.Store(true) }).
		EverySecond().
		Do(func() {})

	waitForJob(clock, 3*time.Second)

	require.True(t, before.Load())
	require.True(t, after.Load())
	require.True(t, success.Load())
}

func TestDaysOfMonth(t *testing.T) {
	f := NewJobBuilder(nil).DaysOfMonth([]int{1, 10, 20}, "13:00")
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

	NewJobBuilder(s).
		Timezone("UTC").
		Weekdays().
		EverySecond().
		Do(func() { weekdayCalled.Store(true) })

	NewJobBuilder(s).
		Timezone("UTC").
		Weekends().
		EverySecond().
		Do(func() { weekendCalled.Store(true) })

	NewJobBuilder(s).
		Timezone("UTC").
		Between("08:00", "10:00").
		EverySecond().
		Do(func() { betweenCalled.Store(true) })

	NewJobBuilder(s).
		Timezone("UTC").
		UnlessBetween("08:00", "10:00").
		EverySecond().
		Do(func() { unlessBetweenSkipped.Store(true) })

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

	NewJobBuilder(s).
		Timezone("UTC").
		Mondays().
		EverySecond().
		Do(func() { mondayCalled.Store(true) })

	NewJobBuilder(s).
		Timezone("UTC").
		Tuesdays().
		EverySecond().
		Do(func() { tuesdayCalled.Store(true) })

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

	NewJobBuilder(s).Timezone("UTC").Sundays().EverySecond().Do(func() { sundayCalled.Store(true) })
	NewJobBuilder(s).Timezone("UTC").Saturdays().EverySecond().Do(func() { saturdayCalled.Store(true) })

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

	NewJobBuilder(s).Timezone("UTC").Wednesdays().EverySecond().Do(func() { wedCalled.Store(true) })
	NewJobBuilder(s).Timezone("UTC").Thursdays().EverySecond().Do(func() { thuCalled.Store(true) })
	NewJobBuilder(s).Timezone("UTC").Fridays().EverySecond().Do(func() { friCalled.Store(true) })

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

	f := NewJobBuilder(s).
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
	f := NewJobBuilder(nil).DaysOfMonth([]int{1, 2}, "bad")
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

	NewJobBuilder(s).Timezone("UTC").Between("23:00", "02:00").EverySecond().Do(func() { betweenCalled.Store(true) })
	NewJobBuilder(s).Timezone("UTC").UnlessBetween("23:00", "02:00").EverySecond().Do(func() { unlessBetweenCalled.Store(true) })

	waitForJob(clock, time.Second)

	require.True(t, betweenCalled.Load())
	require.False(t, unlessBetweenCalled.Load())
}

func TestAddWhenSkipComposition(t *testing.T) {
	f := NewJobBuilder(nil)
	f.When(func() bool { return true })
	f.When(func() bool { return false })
	require.False(t, f.whenFunc())

	f.Skip(func() bool { return false })
	f.Skip(func() bool { return true })
	require.True(t, f.skipFunc())
}
