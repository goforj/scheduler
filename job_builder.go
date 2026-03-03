package scheduler

import (
	"context"
	"fmt"
	"github.com/goforj/cache"
	"github.com/redis/go-redis/v9"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

var nowFunc = time.Now

// CommandRunner abstracts exec so callers/tests can override.
type CommandRunner interface {
	Run(ctx context.Context, exe string, args []string) error
}

// CommandRunnerFunc adapts a function to satisfy CommandRunner.
// @group Adapters
//
// Example: wrap a function as a runner
//
//	runner := scheduler.CommandRunnerFunc(func(ctx context.Context, exe string, args []string) error {
//		return nil
//	})
//	_ = runner.Run(context.Background(), "/bin/true", []string{})
type CommandRunnerFunc func(ctx context.Context, exe string, args []string) error

// Run executes the underlying function.
// @group Adapters
//
// Example: execute the wrapped function
//
//	runner := scheduler.CommandRunnerFunc(func(ctx context.Context, exe string, args []string) error {
//		return nil
//	})
//	_ = runner.Run(context.Background(), "echo", []string{"hi"})
func (f CommandRunnerFunc) Run(ctx context.Context, exe string, args []string) error {
	return f(ctx, exe, args)
}

// SchedulerAdapter wraps gocron.Scheduler with a small interface surface.
type SchedulerAdapter interface {
	NewJob(job gocron.JobDefinition, task gocron.Task, options ...gocron.JobOption) (gocron.Job, error)
	Start()
	Shutdown() error
	Jobs() []gocron.Job
}

type gocronSchedulerAdapter struct {
	s gocron.Scheduler
}

func (a gocronSchedulerAdapter) NewJob(job gocron.JobDefinition, task gocron.Task, options ...gocron.JobOption) (gocron.Job, error) {
	return a.s.NewJob(job, task, options...)
}

func (a gocronSchedulerAdapter) Start() { a.s.Start() }

func (a gocronSchedulerAdapter) Shutdown() error { return a.s.Shutdown() }

func (a gocronSchedulerAdapter) Jobs() []gocron.Job { return a.s.Jobs() }

type execCommandRunner struct{}

func (execCommandRunner) Run(ctx context.Context, exe string, args []string) error {
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// JobBuilder is a wrapper around gocron.Job that provides a fluent interface for scheduling jobs.
type JobBuilder struct {
	scheduler         SchedulerAdapter
	state             *runtimeState
	timezone          string
	name              string
	err               error
	duration          *time.Duration
	cronExpr          string
	withoutOverlap    bool
	retainState       bool
	lastJob           gocron.Job
	distributedLocker gocron.Locker
	envs              []string
	whenFunc          func() bool
	skipFunc          func() bool
	extraTags         []string
	targetKind        jobTargetKind
	commandArgs       []string
	commandRunner     CommandRunner
	now               func() time.Time
	hooks             taskHooks
	runInBackground   bool
}

type taskHooks struct {
	Before    func()
	After     func()
	OnSuccess func()
	OnFailure func()
}

func newJobBuilder(s SchedulerAdapter) *JobBuilder {
	return newJobBuilderWithState(s, nil)
}

func newJobBuilderWithState(s SchedulerAdapter, state *runtimeState) *JobBuilder {
	if state == nil {
		state = newRuntimeState()
	}
	return &JobBuilder{
		scheduler:     s,
		state:         state,
		targetKind:    jobTargetFunction,
		commandRunner: execCommandRunner{},
		now:           nowFunc,
	}
}

func NewJobBuilder(s SchedulerAdapter) *JobBuilder {
	return newJobBuilder(s)
}

type jobScheduleKind string

const (
	jobScheduleCron     jobScheduleKind = "cron"
	jobScheduleInterval jobScheduleKind = "interval"
	jobScheduleUnknown  jobScheduleKind = "unknown"
)

type jobTargetKind string

const (
	jobTargetFunction jobTargetKind = "function"
	jobTargetCommand  jobTargetKind = "command"
	jobTargetExec     jobTargetKind = "exec"
)

// JobMetadata captures stored job details keyed by job ID.
type JobMetadata struct {
	ID           uuid.UUID
	Name         string
	Schedule     string
	ScheduleType string
	TargetKind   string
	Handler      string
	Command      string
	Tags         []string
	Paused       bool
}

// RetainState allows the job to retain its state after execution.
// @group State management
//
// Example: reuse interval configuration for multiple jobs
//
//	builder := scheduler.New().EverySecond().RetainState()
//	builder.Do(func() {})
//	builder.Do(func() {})
func (j *JobBuilder) RetainState() *JobBuilder {
	j.retainState = true
	return j
}

// buildTags constructs tags for the job based on its configuration.
func (j *JobBuilder) buildTags(jobName string) []string {
	var tags []string

	// Always include environment
	currentEnv := getEnv("APP_ENV", "local")
	tags = append(tags, "env="+currentEnv)

	// Interval / cron info
	if j.cronExpr != "" {
		tags = append(tags, "cron="+j.cronExpr)
	}
	if j.duration != nil {
		tags = append(tags, "interval="+j.duration.String())
	}

	// Job name
	if jobName != "" {
		tags = append(tags, "name="+jobName)
	}

	// Args/flags provided via Command/Exec.
	if len(j.extraTags) > 0 {
		tags = append(tags, j.extraTags...)
	}

	return tags
}

// Do schedules the job with the provided task function.
// @group Triggers
//
// Example: create a named cron job
//
//	scheduler.New().Name("cleanup").Cron("0 0 * * *").Do(func() {})
func (j *JobBuilder) Do(task func()) *JobBuilder {
	return j.doWithRunner(task, func() error {
		task()
		return nil
	})
}

func (j *JobBuilder) doWithRunner(task func(), run func() error) *JobBuilder {
	if j.err != nil {
		return j
	}

	localHooks := j.hooks
	bg := j.runInBackground

	if j.targetKind == "" {
		j.targetKind = jobTargetFunction
	}

	var jobName string
	if j.name != "" {
		jobName = j.name
	}

	var opts []gocron.JobOption
	if j.withoutOverlap {
		opts = append(opts, gocron.WithSingletonMode(gocron.LimitModeReschedule))
	}
	if j.distributedLocker != nil {
		opts = append(opts, gocron.WithDistributedJobLocker(j.distributedLocker))
	}
	if jobName != "" {
		opts = append(opts, gocron.WithName(jobName))
	}

	if len(j.envs) > 0 {
		current := getEnv("APP_ENV", "local")
		matched := false
		for _, e := range j.envs {
			if e == current {
				matched = true
				break
			}
		}
		if !matched {
			return j // skip scheduling this job in non-matching environment
		}
	}

	if j.whenFunc != nil && !j.whenFunc() {
		return j // don't job if condition fails
	}
	if j.skipFunc != nil && j.skipFunc() {
		return j // don't job if skip triggers
	}

	// Build tags
	tags := j.buildTags(jobName)
	if len(tags) > 0 {
		opts = append(opts, gocron.WithTags(tags...))
	}

	var jobID uuid.UUID
	taskToRun := func() {
		j.executeRun(jobID, run, localHooks, bg)
	}

	var job gocron.Job
	if j.cronExpr != "" {
		expr := j.cronExpr
		if j.timezone != "" && !strings.HasPrefix(expr, "CRON_TZ=") && !strings.HasPrefix(expr, "TZ=") {
			expr = fmt.Sprintf("CRON_TZ=%s %s", j.timezone, expr)
		}

		job, j.err = j.scheduler.NewJob(
			gocron.CronJob(expr, true),
			gocron.NewTask(taskToRun),
			opts...,
		)
		if j.err != nil {
			j.err = fmt.Errorf("failed to create job with cron expression %q: %w", expr, j.err)
			return j
		}
	} else if j.duration != nil {
		job, j.err = j.scheduler.NewJob(
			gocron.DurationJob(*j.duration),
			gocron.NewTask(taskToRun),
			opts...,
		)
		if j.err != nil {
			return j
		}
	} else {
		j.err = fmt.Errorf("no job defined before calling Do")
		return j
	}

	jobID = job.ID()
	j.lastJob = job
	j.recordJob(job, task)

	// Graceful reset unless .RetainState() was called
	if !j.retainState {
		j.resetState()
	} else {
		j.retainState = false // reset after honoring once
		j.commandArgs = nil
		j.targetKind = jobTargetFunction
	}

	return j
}

func (j *JobBuilder) resetState() {
	j.duration = nil
	j.cronExpr = ""
	j.withoutOverlap = false
	j.distributedLocker = nil
	j.envs = nil
	j.whenFunc = nil
	j.skipFunc = nil
	j.commandArgs = nil
	j.targetKind = jobTargetFunction
	j.name = ""
	j.hooks = taskHooks{}
	j.runInBackground = false
}

// WithoutOverlapping ensures the job does not run concurrently.
// @group Concurrency
//
// Example: prevent overlapping runs of a slow task
//
//	scheduler.New().
//		WithoutOverlapping().
//		EveryFiveSeconds().
//		Do(func() { time.Sleep(7 * time.Second) })
func (j *JobBuilder) WithoutOverlapping() *JobBuilder {
	j.withoutOverlap = true
	return j
}

// Error returns the error if any occurred during job scheduling.
// @group Diagnostics
//
// Example: validate a malformed schedule
//
//	builder := scheduler.New().DailyAt("bad")
//	fmt.Println(builder.Error())
//	// Output: invalid DailyAt time format: invalid time format (expected HH:MM): "bad"
func (j *JobBuilder) Error() error {
	return j.err
}

// Cron sets the cron expression for the job.
// @group Triggers
//
// Example: configure a cron expression
//
//	builder := scheduler.New().Cron("15 3 * * *")
//	fmt.Println(builder.CronExpr())
//	// Output: 15 3 * * *
func (j *JobBuilder) Cron(expr string) *JobBuilder {
	j.cronExpr = expr
	return j
}

// every sets the duration for the job to run at regular intervals.
// @group Intervals
//
// Example: build a duration interval
//
//	_ = scheduler.New().Every(5).Seconds()
//	_ = time.Second
func (j *JobBuilder) every(duration time.Duration) *JobBuilder {
	j.duration = &duration
	return j
}

// Every schedules a job to run every X seconds, minutes, or hours.
// @group Intervals
//
// Example: fluently choose an interval
//
//	scheduler.New().Every(10).Minutes()
func (j *JobBuilder) Every(duration int) *FluentEvery {
	return &FluentEvery{
		base:     j,
		interval: duration,
	}
}

// FluentEvery is a wrapper for scheduling jobs at regular intervals.
type FluentEvery struct {
	interval int
	base     *JobBuilder
}

// Seconds schedules the job to run every X seconds.
// @group Intervals
//
// Example: run a task every few seconds
//
//	scheduler.New().Every(3).Seconds().Do(func() {})
func (fe *FluentEvery) Seconds() *JobBuilder {
	return fe.base.every(time.Duration(fe.interval) * time.Second)
}

// Minutes schedules the job to run every X minutes.
// @group Intervals
//
// Example: chain a minute-based interval
//
//	scheduler.New().Every(15).Minutes()
func (fe *FluentEvery) Minutes() *JobBuilder {
	return fe.base.every(time.Duration(fe.interval) * time.Minute)
}

// Hours schedules the job to run every X hours.
// @group Intervals
//
// Example: build an hourly cadence
//
//	scheduler.New().Every(6).Hours()
func (fe *FluentEvery) Hours() *JobBuilder {
	return fe.base.every(time.Duration(fe.interval) * time.Hour)
}

// EverySecond schedules the job to run every 1 second.
// @group Intervals
//
// Example: heartbeat job each second
//
//	scheduler.New().EverySecond().Do(func() {})
func (j *JobBuilder) EverySecond() *JobBuilder {
	return j.every(1 * time.Second)
}

// EveryTwoSeconds schedules the job to run every 2 seconds.
// @group Intervals
//
// Example: throttle a task to two seconds
//
//	scheduler.New().EveryTwoSeconds().Do(func() {})
func (j *JobBuilder) EveryTwoSeconds() *JobBuilder {
	return j.every(2 * time.Second)
}

// EveryFiveSeconds schedules the job to run every 5 seconds.
// @group Intervals
//
// Example: space out work every five seconds
//
//	scheduler.New().EveryFiveSeconds().Do(func() {})
func (j *JobBuilder) EveryFiveSeconds() *JobBuilder {
	return j.every(5 * time.Second)
}

// EveryTenSeconds schedules the job to run every 10 seconds.
// @group Intervals
//
// Example: poll every ten seconds
//
//	scheduler.New().EveryTenSeconds().Do(func() {})
func (j *JobBuilder) EveryTenSeconds() *JobBuilder {
	return j.every(10 * time.Second)
}

// EveryFifteenSeconds schedules the job to run every 15 seconds.
// @group Intervals
//
// Example: run at fifteen-second cadence
//
//	scheduler.New().EveryFifteenSeconds().Do(func() {})
func (j *JobBuilder) EveryFifteenSeconds() *JobBuilder {
	return j.every(15 * time.Second)
}

// EveryTwentySeconds schedules the job to run every 20 seconds.
// @group Intervals
//
// Example: run once every twenty seconds
//
//	scheduler.New().EveryTwentySeconds().Do(func() {})
func (j *JobBuilder) EveryTwentySeconds() *JobBuilder {
	return j.every(20 * time.Second)
}

// EveryThirtySeconds schedules the job to run every 30 seconds.
// @group Intervals
//
// Example: execute every thirty seconds
//
//	scheduler.New().EveryThirtySeconds().Do(func() {})
func (j *JobBuilder) EveryThirtySeconds() *JobBuilder {
	return j.every(30 * time.Second)
}

// EveryMinute schedules the job to run every 1 minute.
// @group Intervals
//
// Example: run a task each minute
//
//	scheduler.New().EveryMinute().Do(func() {})
func (j *JobBuilder) EveryMinute() *JobBuilder {
	return j.every(1 * time.Minute)
}

// EveryTwoMinutes schedules the job to run every 2 minutes.
// @group Intervals
//
// Example: job that runs every two minutes
//
//	scheduler.New().EveryTwoMinutes().Do(func() {})
func (j *JobBuilder) EveryTwoMinutes() *JobBuilder {
	return j.every(2 * time.Minute)
}

// EveryThreeMinutes schedules the job to run every 3 minutes.
// @group Intervals
//
// Example: run every three minutes
//
//	scheduler.New().EveryThreeMinutes().Do(func() {})
func (j *JobBuilder) EveryThreeMinutes() *JobBuilder {
	return j.every(3 * time.Minute)
}

// EveryFourMinutes schedules the job to run every 4 minutes.
// @group Intervals
//
// Example: run every four minutes
//
//	scheduler.New().EveryFourMinutes().Do(func() {})
func (j *JobBuilder) EveryFourMinutes() *JobBuilder {
	return j.every(4 * time.Minute)
}

// EveryFiveMinutes schedules the job to run every 5 minutes.
// @group Intervals
//
// Example: run every five minutes
//
//	scheduler.New().EveryFiveMinutes().Do(func() {})
func (j *JobBuilder) EveryFiveMinutes() *JobBuilder {
	return j.every(5 * time.Minute)
}

// EveryTenMinutes schedules the job to run every 10 minutes.
// @group Intervals
//
// Example: run every ten minutes
//
//	scheduler.New().EveryTenMinutes().Do(func() {})
func (j *JobBuilder) EveryTenMinutes() *JobBuilder {
	return j.every(10 * time.Minute)
}

// EveryFifteenMinutes schedules the job to run every 15 minutes.
// @group Intervals
//
// Example: run every fifteen minutes
//
//	scheduler.New().EveryFifteenMinutes().Do(func() {})
func (j *JobBuilder) EveryFifteenMinutes() *JobBuilder {
	return j.every(15 * time.Minute)
}

// EveryThirtyMinutes schedules the job to run every 30 minutes.
// @group Intervals
//
// Example: run every thirty minutes
//
//	scheduler.New().EveryThirtyMinutes().Do(func() {})
func (j *JobBuilder) EveryThirtyMinutes() *JobBuilder {
	return j.every(30 * time.Minute)
}

// Hourly schedules the job to run every hour.
// @group Intervals
//
// Example: run something hourly
//
//	scheduler.New().Hourly().Do(func() {})
func (j *JobBuilder) Hourly() *JobBuilder {
	return j.every(1 * time.Hour)
}

// HourlyAt schedules the job to run every hour at the specified minute.
// @group Intervals
//
// Example: run at the 5th minute of each hour
//
//	scheduler.New().HourlyAt(5)
func (j *JobBuilder) HourlyAt(minute int) *JobBuilder {
	return j.Cron(fmt.Sprintf("%d * * * *", minute))
}

// EveryOddHour schedules the job to run every odd-numbered hour at the specified minute.
// @group Intervals
//
// Example: run every odd hour
//
//	scheduler.New().EveryOddHour(10)
func (j *JobBuilder) EveryOddHour(minute int) *JobBuilder {
	return j.Cron(fmt.Sprintf("%d 1-23/2 * * *", minute))
}

// EveryTwoHours schedules the job to run every two hours at the specified minute.
// @group Intervals
//
// Example: run every two hours
//
//	scheduler.New().EveryTwoHours(15)
func (j *JobBuilder) EveryTwoHours(minute int) *JobBuilder {
	return j.Cron(fmt.Sprintf("%d */2 * * *", minute))
}

// EveryThreeHours schedules the job to run every three hours at the specified minute.
// @group Intervals
//
// Example: run every three hours
//
//	scheduler.New().EveryThreeHours(20)
func (j *JobBuilder) EveryThreeHours(minute int) *JobBuilder {
	return j.Cron(fmt.Sprintf("%d */3 * * *", minute))
}

// EveryFourHours schedules the job to run every four hours at the specified minute.
// @group Intervals
//
// Example: run every four hours
//
//	scheduler.New().EveryFourHours(25)
func (j *JobBuilder) EveryFourHours(minute int) *JobBuilder {
	return j.Cron(fmt.Sprintf("%d */4 * * *", minute))
}

// EverySixHours schedules the job to run every six hours at the specified minute.
// @group Intervals
//
// Example: run every six hours
//
//	scheduler.New().EverySixHours(30)
func (j *JobBuilder) EverySixHours(minute int) *JobBuilder {
	return j.Cron(fmt.Sprintf("%d */6 * * *", minute))
}

// Daily schedules the job to run once per day at midnight.
// @group Calendar
//
// Example: nightly task
//
//	scheduler.New().Daily()
func (j *JobBuilder) Daily() *JobBuilder {
	return j.Cron("0 0 * * *")
}

// DailyAt schedules the job to run daily at a specific time (e.g., "13:00").
// @group Calendar
//
// Example: run at lunch time daily
//
//	scheduler.New().DailyAt("12:30")
func (j *JobBuilder) DailyAt(hm string) *JobBuilder {
	hour, minute, err := parseHourMinute(hm)
	if err != nil {
		j.err = fmt.Errorf("invalid DailyAt format: %w", err)
		return j
	}
	return j.Cron(fmt.Sprintf("%d %d * * *", minute, hour))
}

// TwiceDaily schedules the job to run daily at two specified hours (e.g., 1 and 13).
// @group Calendar
//
// Example: run two times per day
//
//	scheduler.New().TwiceDaily(1, 13)
func (j *JobBuilder) TwiceDaily(h1, h2 int) *JobBuilder {
	return j.Cron(fmt.Sprintf("0 %d,%d * * *", h1, h2))
}

// TwiceDailyAt schedules the job to run daily at two specified times (e.g., 1:15 and 13:15).
// @group Calendar
//
// Example: run twice daily at explicit minutes
//
//	scheduler.New().TwiceDailyAt(1, 13, 15)
func (j *JobBuilder) TwiceDailyAt(h1, h2, m int) *JobBuilder {
	return j.Cron(fmt.Sprintf("%d %d,%d * * *", m, h1, h2))
}

// Weekly schedules the job to run once per week on Sunday at midnight.
// @group Calendar
//
// Example: weekly maintenance
//
//	scheduler.New().Weekly()
func (j *JobBuilder) Weekly() *JobBuilder {
	return j.Cron("0 0 * * 0")
}

// WeeklyOn schedules the job to run weekly on a specific day of the week and time.
// Day uses 0 = Sunday through 6 = Saturday.
// @group Calendar
//
// Example: run each Monday at 08:00
//
//	scheduler.New().WeeklyOn(1, "8:00")
func (j *JobBuilder) WeeklyOn(day int, hm string) *JobBuilder {
	parts := strings.Split(hm, ":")
	if len(parts) != 2 {
		j.err = fmt.Errorf("invalid WeeklyOn format: expected HH:MM but got %s", hm)
		return j
	}
	hourStr := strings.TrimSpace(parts[0])
	minuteStr := strings.TrimSpace(parts[1])

	hour, err := strconv.Atoi(hourStr)
	if err != nil {
		j.err = fmt.Errorf("invalid hour in WeeklyOn time string %q: %w", hourStr, err)
		return j
	}
	minute, err := strconv.Atoi(minuteStr)
	if err != nil {
		j.err = fmt.Errorf("invalid minute in WeeklyOn time string %q: %w", minuteStr, err)
		return j
	}

	return j.Cron(fmt.Sprintf("%d %d * * %d", minute, hour, day))
}

// Monthly schedules the job to run on the first day of each month at midnight.
// @group Calendar
//
// Example: first-of-month billing
//
//	scheduler.New().Monthly()
func (j *JobBuilder) Monthly() *JobBuilder {
	return j.Cron("0 0 1 * *")
}

// MonthlyOn schedules the job to run on a specific day of the month at a given time.
// @group Calendar
//
// Example: run on the 15th of each month
//
//	scheduler.New().MonthlyOn(15, "09:30")
func (j *JobBuilder) MonthlyOn(day int, hm string) *JobBuilder {
	hour, minute, err := parseHourMinute(hm)
	if err != nil {
		j.err = fmt.Errorf("invalid MonthlyOn time format: %w", err)
		return j
	}
	return j.Cron(fmt.Sprintf("%d %d %d * *", minute, hour, day))
}

// TwiceMonthly schedules the job to run on two specific days of the month at the given time.
// @group Calendar
//
// Example: run on two days each month
//
//	scheduler.New().TwiceMonthly(1, 15, "10:00")
func (j *JobBuilder) TwiceMonthly(d1, d2 int, hm string) *JobBuilder {
	hour, minute, err := parseHourMinute(hm)
	if err != nil {
		j.err = fmt.Errorf("invalid TwiceMonthly time format: %w", err)
		return j
	}
	return j.Cron(fmt.Sprintf("%d %d %d,%d * *", minute, hour, d1, d2))
}

// LastDayOfMonth schedules the job to run on the last day of each month at a specific time.
// @group Calendar
//
// Example: run on the last day of the month
//
//	scheduler.New().LastDayOfMonth("23:30")
func (j *JobBuilder) LastDayOfMonth(hm string) *JobBuilder {
	hour, minute, err := parseHourMinute(hm)
	if err != nil {
		j.err = fmt.Errorf("invalid LastDayOfMonth time format: %w", err)
		return j
	}
	return j.Cron(fmt.Sprintf("%d %d L * *", minute, hour))
}

// Quarterly schedules the job to run on the first day of each quarter at midnight.
// @group Calendar
//
// Example: quarterly trigger
//
//	scheduler.New().Quarterly()
func (j *JobBuilder) Quarterly() *JobBuilder {
	return j.Cron("0 0 1 1,4,7,10 *")
}

// QuarterlyOn schedules the job to run on a specific day of each quarter at a given time.
// @group Calendar
//
// Example: quarterly on a specific day
//
//	scheduler.New().QuarterlyOn(3, "12:00")
func (j *JobBuilder) QuarterlyOn(day int, hm string) *JobBuilder {
	hour, minute, err := parseHourMinute(hm)
	if err != nil {
		j.err = fmt.Errorf("invalid QuarterlyOn time format: %w", err)
		return j
	}
	return j.Cron(fmt.Sprintf("%d %d %d 1,4,7,10 *", minute, hour, day))
}

// Yearly schedules the job to run on January 1st every year at midnight.
// @group Calendar
//
// Example: yearly trigger
//
//	scheduler.New().Yearly()
func (j *JobBuilder) Yearly() *JobBuilder {
	return j.Cron("0 0 1 1 *")
}

// YearlyOn schedules the job to run every year on a specific month, day, and time.
// @group Calendar
//
// Example: yearly on a specific date
//
//	scheduler.New().YearlyOn(12, 25, "06:45")
func (j *JobBuilder) YearlyOn(month, day int, hm string) *JobBuilder {
	hour, minute, err := parseHourMinute(hm)
	if err != nil {
		j.err = fmt.Errorf("invalid YearlyOn time format: %w", err)
		return j
	}
	return j.Cron(fmt.Sprintf("%d %d %d %d *", minute, hour, day, month))
}

// Timezone sets a timezone string for the job (not currently applied to gocron Scheduler).
// @group Configuration
//
// Example: tag jobs with a timezone
//
//	scheduler.New().Timezone("America/New_York").Daily()
func (j *JobBuilder) Timezone(zone string) *JobBuilder {
	j.timezone = zone
	return j
}

// WithCommandRunner overrides command execution (default: exec.CommandContext).
// @group Configuration
//
// Example: swap in a custom runner
//
//	runner := scheduler.CommandRunnerFunc(func(_ context.Context, exe string, args []string) error {
//		_ = exe
//		_ = args
//		return nil
//	})
//
//	builder := scheduler.New().WithCommandRunner(runner)
//	_ = builder
func (j *JobBuilder) WithCommandRunner(r CommandRunner) *JobBuilder {
	if r != nil {
		j.commandRunner = r
	}
	return j
}

// WithNowFunc overrides current time (default: time.Now). Useful for tests.
// @group Configuration
//
// Example: freeze time for predicates
//
//	fixed := func() time.Time { return time.Unix(0, 0) }
//	scheduler.New().WithNowFunc(fixed)
func (j *JobBuilder) WithNowFunc(fn func() time.Time) *JobBuilder {
	if fn != nil {
		j.now = fn
	}
	return j
}

// DaysOfMonth schedules the job to run on specific days of the month at a given time.
// @group Calendar
//
// Example: run on the 5th and 20th of each month
//
//	scheduler.New().DaysOfMonth([]int{5, 20}, "07:15")
func (j *JobBuilder) DaysOfMonth(days []int, hm string) *JobBuilder {
	hour, minute, err := parseHourMinute(hm)
	if err != nil {
		j.err = fmt.Errorf("invalid DaysOfMonth time format: %w", err)
		return j
	}
	var parts []string
	for _, d := range days {
		parts = append(parts, strconv.Itoa(d))
	}
	return j.Cron(fmt.Sprintf("%d %d %s * *", minute, hour, strings.Join(parts, ",")))
}

// CronExpr returns the cron expression string configured for this job.
// @group Diagnostics
//
// Example: inspect the stored cron expression
//
//	builder := scheduler.New().Cron("0 9 * * *")
//	fmt.Println(builder.CronExpr())
//	// Output: 0 9 * * *
func (j *JobBuilder) CronExpr() string {
	return j.cronExpr
}

// Job returns the last scheduled gocron.Job instance, if available.
// @group Diagnostics
//
// Example: capture the last job handle
//
//	b := scheduler.New().EverySecond().Do(func() {})
//	fmt.Println(b.Job() != nil)
//	// Output: true
func (j *JobBuilder) Job() gocron.Job {
	return j.lastJob
}

// WithoutOverlappingWithLocker ensures the job does not run concurrently across distributed systems using the provided locker.
// @group Concurrency
//
// Example: use a distributed locker
//
//	locker := scheduler.LockerFunc(func(ctx context.Context, key string) (gocron.Lock, error) {
//		return scheduler.LockFunc(func(context.Context) error { return nil }), nil
//	})
//
//	scheduler.New().
//		WithoutOverlappingWithLocker(locker).
//		EveryMinute().
//		Do(func() {})
func (j *JobBuilder) WithoutOverlappingWithLocker(locker gocron.Locker) *JobBuilder {
	j.withoutOverlap = true
	j.distributedLocker = locker
	return j
}

// Environments restricts job registration to specific environment names (e.g. "production", "staging").
// @group Filters
//
// Example: only register in production
//
//	scheduler.New().Environments("production").Daily()
func (j *JobBuilder) Environments(envs ...string) *JobBuilder {
	j.envs = envs
	return j
}

// When only schedules the job if the provided condition returns true.
// @group Filters
//
// Example: guard scheduling with a flag
//
//	flag := true
//	scheduler.New().When(func() bool { return flag }).Daily()
func (j *JobBuilder) When(fn func() bool) *JobBuilder {
	j.addWhen(fn)
	return j
}

// Skip prevents scheduling the job if the provided condition returns true.
// @group Filters
//
// Example: suppress jobs based on a switch
//
//	enabled := false
//	scheduler.New().Skip(func() bool { return !enabled }).Daily()
func (j *JobBuilder) Skip(fn func() bool) *JobBuilder {
	j.addSkip(fn)
	return j
}

// Name sets an explicit job name.
// @group Metadata
//
// Example: label a job for logging
//
//	scheduler.New().Name("cache:refresh").HourlyAt(15)
func (j *JobBuilder) Name(name string) *JobBuilder {
	j.name = name
	return j
}

// RunInBackground runs command/exec tasks in a goroutine.
// @group Execution
//
// Example: allow command jobs to run async
//
//	scheduler.New().RunInBackground().Command("noop")
func (j *JobBuilder) RunInBackground() *JobBuilder {
	j.runInBackground = true
	return j
}

// Before sets a hook to run before task execution.
// @group Hooks
//
// Example: add a before hook
//
//	scheduler.New().Before(func() {}).Daily()
func (j *JobBuilder) Before(fn func()) *JobBuilder {
	j.hooks.Before = fn
	return j
}

// After sets a hook to run after task execution.
// @group Hooks
//
// Example: add an after hook
//
//	scheduler.New().After(func() {}).Daily()
func (j *JobBuilder) After(fn func()) *JobBuilder {
	j.hooks.After = fn
	return j
}

// OnSuccess sets a hook to run after successful task execution.
// @group Hooks
//
// Example: record success
//
//	scheduler.New().OnSuccess(func() {}).Daily()
func (j *JobBuilder) OnSuccess(fn func()) *JobBuilder {
	j.hooks.OnSuccess = fn
	return j
}

// OnFailure sets a hook to run after failed task execution.
// @group Hooks
//
// Example: record failures
//
//	scheduler.New().OnFailure(func() {}).Daily()
func (j *JobBuilder) OnFailure(fn func()) *JobBuilder {
	j.hooks.OnFailure = fn
	return j
}

func (j *JobBuilder) addWhen(fn func() bool) {
	if fn == nil {
		return
	}
	if j.whenFunc == nil {
		j.whenFunc = fn
		return
	}
	prev := j.whenFunc
	j.whenFunc = func() bool {
		return prev() && fn()
	}
}

func (j *JobBuilder) addSkip(fn func() bool) {
	if fn == nil {
		return
	}
	if j.skipFunc == nil {
		j.skipFunc = fn
		return
	}
	prev := j.skipFunc
	j.skipFunc = func() bool {
		return prev() || fn()
	}
}

// Weekdays limits the job to run only on weekdays (Mon-Fri).
// @group Filters
//
// Example: weekday-only execution
//
//	scheduler.New().Weekdays().DailyAt("09:00")
func (j *JobBuilder) Weekdays() *JobBuilder {
	j.addWhen(func() bool {
		return isWeekday(j.now(), j.location())
	})
	return j
}

// Weekends limits the job to run only on weekends (Sat-Sun).
// @group Filters
//
// Example: weekend-only execution
//
//	scheduler.New().Weekends().DailyAt("10:00")
func (j *JobBuilder) Weekends() *JobBuilder {
	j.addWhen(func() bool {
		return isWeekend(j.now(), j.location())
	})
	return j
}

// Sundays limits the job to Sundays.
// @group Filters
//
// Example: run only on Sundays
//
//	scheduler.New().Sundays().DailyAt("09:00")
func (j *JobBuilder) Sundays() *JobBuilder { return j.days(time.Sunday) }

// Mondays limits the job to Mondays.
// @group Filters
//
// Example: run only on Mondays
//
//	scheduler.New().Mondays().DailyAt("09:00")
func (j *JobBuilder) Mondays() *JobBuilder { return j.days(time.Monday) }

// Tuesdays limits the job to Tuesdays.
// @group Filters
//
// Example: run only on Tuesdays
//
//	scheduler.New().Tuesdays().DailyAt("09:00")
func (j *JobBuilder) Tuesdays() *JobBuilder { return j.days(time.Tuesday) }

// Wednesdays limits the job to Wednesdays.
// @group Filters
//
// Example: run only on Wednesdays
//
//	scheduler.New().Wednesdays().DailyAt("09:00")
func (j *JobBuilder) Wednesdays() *JobBuilder { return j.days(time.Wednesday) }

// Thursdays limits the job to Thursdays.
// @group Filters
//
// Example: run only on Thursdays
//
//	scheduler.New().Thursdays().DailyAt("09:00")
func (j *JobBuilder) Thursdays() *JobBuilder { return j.days(time.Thursday) }

// Fridays limits the job to Fridays.
// @group Filters
//
// Example: run only on Fridays
//
//	scheduler.New().Fridays().DailyAt("09:00")
func (j *JobBuilder) Fridays() *JobBuilder { return j.days(time.Friday) }

// Saturdays limits the job to Saturdays.
// @group Filters
//
// Example: run only on Saturdays
//
//	scheduler.New().Saturdays().DailyAt("09:00")
func (j *JobBuilder) Saturdays() *JobBuilder { return j.days(time.Saturday) }

// Days limits the job to a specific set of weekdays.
// @group Filters
//
// Example: pick custom weekdays
//
//	scheduler.New().Days(time.Monday, time.Wednesday, time.Friday).DailyAt("07:00")
func (j *JobBuilder) Days(days ...time.Weekday) *JobBuilder {
	set := make(map[time.Weekday]struct{}, len(days))
	for _, d := range days {
		set[d] = struct{}{}
	}
	j.addWhen(func() bool {
		now := j.now().In(j.location())
		_, ok := set[now.Weekday()]
		return ok
	})
	return j
}

func (j *JobBuilder) days(day time.Weekday) *JobBuilder {
	return j.Days(day)
}

// Between limits the job to run between the provided HH:MM times (inclusive).
// @group Filters
//
// Example: allow execution during business hours
//
//	scheduler.New().Between("09:00", "17:00").EveryMinute()
func (j *JobBuilder) Between(start, end string) *JobBuilder {
	startH, startM, err := parseHourMinute(start)
	if err != nil {
		j.err = fmt.Errorf("invalid Between start time: %w", err)
		return j
	}
	endH, endM, err := parseHourMinute(end)
	if err != nil {
		j.err = fmt.Errorf("invalid Between end time: %w", err)
		return j
	}
	j.addWhen(func() bool {
		loc := j.location()
		now := j.now().In(loc)
		return timeInRange(now, startH, startM, endH, endM)
	})
	return j
}

// UnlessBetween prevents the job from running between the provided HH:MM times.
// @group Filters
//
// Example: pause execution overnight
//
//	scheduler.New().UnlessBetween("22:00", "06:00").EveryMinute()
func (j *JobBuilder) UnlessBetween(start, end string) *JobBuilder {
	startH, startM, err := parseHourMinute(start)
	if err != nil {
		j.err = fmt.Errorf("invalid UnlessBetween start time: %w", err)
		return j
	}
	endH, endM, err := parseHourMinute(end)
	if err != nil {
		j.err = fmt.Errorf("invalid UnlessBetween end time: %w", err)
		return j
	}
	j.addSkip(func() bool {
		loc := j.location()
		now := j.now().In(loc)
		return timeInRange(now, startH, startM, endH, endM)
	})
	return j
}

// Command executes the current binary with the given subcommand and variadic args.
// It does not run arbitrary system executables; use Exec for that.
// @group Commands
//
// Example: run a CLI subcommand on schedule
//
//	scheduler.New().Cron("0 0 * * *").Command("jobs:purge", "--force")
func (j *JobBuilder) Command(subcommand string, args ...string) *JobBuilder {
	return j.scheduleExecTarget(subcommand, args, true, jobTargetCommand)
}

// Exec runs an external executable with variadic args.
// @group Commands
//
// Example: run a system executable on schedule
//
//	scheduler.New().Cron("0 0 * * *").Exec("/usr/bin/env", "echo", "hello")
func (j *JobBuilder) Exec(executable string, args ...string) *JobBuilder {
	return j.scheduleExecTarget(executable, args, false, jobTargetExec)
}

func (j *JobBuilder) scheduleExecTarget(commandOrExecutable string, args []string, selfBinary bool, kind jobTargetKind) *JobBuilder {
	j.name = commandOrExecutable
	j.targetKind = kind
	j.commandArgs = args

	// turn CLI args into tags
	if len(args) > 0 {
		// collapse into one quoted string
		joined := strings.Join(args, " ")
		j.extraTags = []string{fmt.Sprintf("args=\"%s\"", joined)}
	}

	task := func() {}
	run := func() error {
		if j.err != nil {
			return j.err
		}

		exe := commandOrExecutable
		if selfBinary {
			self, err := os.Executable()
			if err != nil {
				return fmt.Errorf("unable to determine executable path: %w", err)
			}
			exe = self
		}

		ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
		defer cancel()

		cmdArgs := args
		if selfBinary {
			cmdArgs = append([]string{commandOrExecutable}, args...)
		}
		return j.commandRunner.Run(ctx, exe, cmdArgs)
	}

	return j.doWithRunner(task, run)
}

// PauseAll pauses execution for all scheduled jobs without removing them.
// This is universal across Do, Command, and Exec jobs.
// RunNow calls are skipped while pause is active.
// @group Runtime control
//
// Example: pause all jobs
//
//	s := scheduler.New()
//	_ = s.PauseAll()
func (j *JobBuilder) PauseAll() error {
	if j.state == nil {
		j.state = newRuntimeState()
	}
	j.state.pauseAll()
	return nil
}

// ResumeAll resumes execution for all paused jobs.
// @group Runtime control
//
// Example: resume all jobs
//
//	s := scheduler.New()
//	_ = s.ResumeAll()
func (j *JobBuilder) ResumeAll() error {
	if j.state == nil {
		j.state = newRuntimeState()
	}
	j.state.resumeAll()
	return nil
}

// PauseJob pauses execution for a specific scheduled job.
// RunNow calls for that job are skipped while paused.
// @group Runtime control
//
// Example: pause one job by ID
//
//	s := scheduler.New()
//	b := s.EverySecond().Name("heartbeat").Do(func() {})
//	_ = s.PauseJob(b.Job().ID())
func (j *JobBuilder) PauseJob(id uuid.UUID) error {
	if j.state == nil {
		j.state = newRuntimeState()
	}
	j.state.pauseJob(id)
	return nil
}

// ResumeJob resumes a paused job by ID.
// @group Runtime control
//
// Example: resume one job by ID
//
//	s := scheduler.New()
//	b := s.EverySecond().Name("heartbeat").Do(func() {})
//	_ = s.ResumeJob(b.Job().ID())
func (j *JobBuilder) ResumeJob(id uuid.UUID) error {
	if j.state == nil {
		j.state = newRuntimeState()
	}
	j.state.resumeJob(id)
	return nil
}

// IsPausedAll reports whether global pause is enabled.
// @group Runtime control
func (j *JobBuilder) IsPausedAll() bool {
	if j.state == nil {
		return false
	}
	return j.state.isPausedAll()
}

// IsJobPaused reports whether a specific job is paused.
// @group Runtime control
func (j *JobBuilder) IsJobPaused(id uuid.UUID) bool {
	if j.state == nil {
		return false
	}
	return j.state.isJobPaused(id)
}

// Observe registers a lifecycle observer for all scheduled jobs.
// Events are emitted consistently across Do, Command, and Exec jobs.
// @group Runtime control
//
// Example: observe paused-skip events
//
//	s := scheduler.New()
//	s.Observe(scheduler.JobObserverFunc(func(event scheduler.JobEvent) {
//		if event.Type == scheduler.JobSkipped && event.Reason == "paused" {
//			fmt.Println("skipped: paused")
//		}
//	}))
func (j *JobBuilder) Observe(observer JobObserver) *JobBuilder {
	if j.state == nil {
		j.state = newRuntimeState()
	}
	j.state.addObserver(observer)
	return j
}

// JobMetadata returns a copy of the tracked job metadata keyed by job ID.
// @group Metadata
//
// Example: inspect scheduled jobs
//
//	b := scheduler.New().EverySecond().Do(func() {})
//	for id, meta := range b.JobMetadata() {
//		_ = id
//		_ = meta.Name
//	}
func (j *JobBuilder) JobMetadata() map[uuid.UUID]JobMetadata {
	if j.state == nil {
		return map[uuid.UUID]JobMetadata{}
	}
	return j.state.metadataCopy()
}

// JobsInfo returns a stable, sorted snapshot of all known job metadata.
// This is a facade-friendly list form of JobMetadata including paused state.
// @group Metadata
//
// Example: iterate jobs for UI rendering
//
//	s := scheduler.New()
//	s.EverySecond().Name("heartbeat").Do(func() {})
//	for _, job := range s.JobsInfo() {
//		_ = job.ID
//		_ = job.Name
//		_ = job.Paused
//	}
func (j *JobBuilder) JobsInfo() []JobMetadata {
	if j.state == nil {
		return []JobMetadata{}
	}
	return j.state.metadataList()
}

func (j *JobBuilder) recordJob(job gocron.Job, task func()) {
	if j.state == nil {
		j.state = newRuntimeState()
	}

	scheduleType, schedule := j.describeSchedule()
	kind := j.targetKind
	if kind == "" {
		kind = jobTargetFunction
	}

	meta := JobMetadata{
		ID:           job.ID(),
		Name:         job.Name(),
		Schedule:     schedule,
		ScheduleType: string(scheduleType),
		TargetKind:   string(kind),
		Tags:         job.Tags(),
	}

	if kind == jobTargetCommand || kind == jobTargetExec {
		meta.Command = buildCommandString(j.name, j.commandArgs)
		if meta.Name == "" {
			meta.Name = j.name
		}
	} else {
		meta.Handler = friendlyFuncName(task)
		if meta.Name == "" && j.name != "" {
			meta.Name = j.name
		}
	}

	j.state.setMetadata(meta)
}

func (j *JobBuilder) describeSchedule() (jobScheduleKind, string) {
	switch {
	case j.cronExpr != "":
		if j.timezone != "" {
			return jobScheduleCron, fmt.Sprintf("CRON_TZ=%s %s", j.timezone, j.cronExpr)
		}
		return jobScheduleCron, j.cronExpr
	case j.duration != nil:
		return jobScheduleInterval, j.duration.String()
	default:
		return jobScheduleUnknown, ""
	}
}

func buildCommandString(name string, args []string) string {
	if name == "" {
		return ""
	}
	if len(args) == 0 {
		return name
	}
	return strings.TrimSpace(name + " " + strings.Join(args, " "))
}

func friendlyFuncName(fn func()) string {
	if fn == nil {
		return ""
	}

	ptr := reflect.ValueOf(fn).Pointer()
	rf := runtime.FuncForPC(ptr)
	if rf == nil {
		return ""
	}

	name := rf.Name()
	name = strings.TrimSuffix(name, "-fm")

	anon := false
	funcRe := regexp.MustCompile(`\.func\d+`)
	if funcRe.MatchString(name) {
		anon = true
	}
	name = funcRe.ReplaceAllString(name, "")

	name = filepath.ToSlash(name)
	if idx := strings.LastIndex(name, "/"); idx != -1 && idx+1 < len(name) {
		name = name[idx+1:]
	}

	name = strings.ReplaceAll(name, "(*", "")
	name = strings.ReplaceAll(name, ")", "")
	name = strings.ReplaceAll(name, "..", ".")
	name = strings.TrimPrefix(name, ".")

	segments := strings.Split(name, ".")
	if len(segments) >= 3 {
		n := len(segments)
		typePart := segments[n-3]
		method := segments[n-2]
		anonSuffix := ""
		if anon {
			anonSuffix = " (anon func)"
		}
		return strings.TrimSpace(fmt.Sprintf("%s.%s%s", typePart, method, anonSuffix))
	}
	if len(segments) == 2 {
		out := strings.Join(segments, ".")
		if anon {
			return out + " (anon func)"
		}
		return out
	}
	if len(segments) == 1 {
		out := segments[0]
		if anon {
			return out + " (anon func)"
		}
		return out
	}
	return name
}

func (j *JobBuilder) location() *time.Location {
	if j.timezone == "" {
		return time.Local
	}
	loc, err := time.LoadLocation(j.timezone)
	if err != nil {
		return time.Local
	}
	return loc
}

func (j *JobBuilder) executeRun(jobID uuid.UUID, run func() error, hooks taskHooks, bg bool) {
	execFn := func() {
		if j.state != nil && j.state.isExecutionPaused(jobID) {
			meta := j.JobMetadata()[jobID]
			j.state.emit(JobEvent{
				Type:       JobSkipped,
				JobID:      jobID,
				Name:       meta.Name,
				TargetKind: meta.TargetKind,
				Reason:     string(JobSkipPaused),
				OccurredAt: time.Now(),
			})
			return
		}

		meta := j.JobMetadata()[jobID]
		attempt := 1
		if j.state != nil {
			attempt = j.state.nextAttempt(jobID)
		}

		start := time.Now()
		if j.state != nil {
			j.state.emit(JobEvent{
				Type:       JobStarted,
				JobID:      jobID,
				Name:       meta.Name,
				TargetKind: meta.TargetKind,
				Attempt:    attempt,
				OccurredAt: start,
			})
		}

		if hooks.Before != nil {
			hooks.Before()
		}

		err := run()
		duration := time.Since(start)
		if err != nil {
			if hooks.OnFailure != nil {
				hooks.OnFailure()
			}
			if j.state != nil {
				j.state.emit(JobEvent{
					Type:       JobFailed,
					JobID:      jobID,
					Name:       meta.Name,
					TargetKind: meta.TargetKind,
					Attempt:    attempt,
					Duration:   duration,
					Error:      err,
					OccurredAt: time.Now(),
				})
			}
		} else {
			if hooks.OnSuccess != nil {
				hooks.OnSuccess()
			}
			if j.state != nil {
				j.state.emit(JobEvent{
					Type:       JobSucceeded,
					JobID:      jobID,
					Name:       meta.Name,
					TargetKind: meta.TargetKind,
					Attempt:    attempt,
					Duration:   duration,
					OccurredAt: time.Now(),
				})
			}
		}

		if hooks.After != nil {
			hooks.After()
		}
	}

	if bg {
		go execFn()
		return
	}
	execFn()
}

func isWeekday(t time.Time, loc *time.Location) bool {
	w := t.In(loc).Weekday()
	return w >= time.Monday && w <= time.Friday
}

func isWeekend(t time.Time, loc *time.Location) bool {
	w := t.In(loc).Weekday()
	return w == time.Saturday || w == time.Sunday
}

func timeInRange(now time.Time, startH, startM, endH, endM int) bool {
	loc := now.Location()
	start := time.Date(now.Year(), now.Month(), now.Day(), startH, startM, 0, 0, loc)
	end := time.Date(now.Year(), now.Month(), now.Day(), endH, endM, 0, 0, loc)

	if end.Before(start) {
		// crosses midnight: treat as two ranges (start..23:59) or (00:00..end)
		if !now.Before(start) {
			return true
		}
		tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, endH, endM, 0, 0, loc)
		return !now.After(tomorrow)
	}

	return (now.Equal(start) || now.After(start)) && (now.Equal(end) || now.Before(end))
}

func getEnv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

// parseHourMinute parses a string in the format "HH:MM" and returns the hour and minute as integers.
func parseHourMinute(hm string) (int, int, error) {
	parts := strings.Split(hm, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format (expected HH:MM): %q", hm)
	}
	hour, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid hour: %w", err)
	}
	minute, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minute: %w", err)
	}
	return hour, minute, nil
}

// redisLockerClient is the minimal interface we need from a Redis client.
type redisLockerClient interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

// LockerFunc adapts a function to satisfy gocron.Locker.
// @group Adapters
//
// Example: build a locker from a function
//
//	locker := scheduler.LockerFunc(func(ctx context.Context, key string) (gocron.Lock, error) {
//		return scheduler.LockFunc(func(context.Context) error { return nil }), nil
//	})
//	_, _ = locker.Lock(context.Background(), "job")
type LockerFunc func(ctx context.Context, key string) (gocron.Lock, error)

// Lock invokes the underlying function.
// @group Adapters
func (f LockerFunc) Lock(ctx context.Context, key string) (gocron.Lock, error) {
	return f(ctx, key)
}

// LockFunc adapts a function to satisfy gocron.Lock.
// @group Adapters
//
// Example: unlock via a function
//
//	lock := scheduler.LockFunc(func(context.Context) error { return nil })
//	_ = lock.Unlock(context.Background())
type LockFunc func(ctx context.Context) error

// Unlock invokes the underlying function.
// @group Adapters
//
// Example: invoke unlock on an adapted lock
//
//	lock := scheduler.LockFunc(func(context.Context) error { return nil })
//	_ = lock.Unlock(context.Background())
func (f LockFunc) Unlock(ctx context.Context) error {
	return f(ctx)
}

// CacheLocker adapts a cache lock API to gocron.Locker.
// It uses a single key per job and auto-expires after ttl.
type CacheLocker struct {
	client cache.LockAPI
	ttl    time.Duration
}

// NewCacheLocker creates a CacheLocker with a cache lock client and TTL.
// The ttl is a lease duration: when it expires, another worker may acquire the
// same lock key. For long-running jobs, choose ttl >= worst-case runtime plus a
// safety buffer. If your runtime can exceed ttl, prefer a renewing/heartbeat lock strategy.
// @group Locking
//
// Example: use an in-memory cache driver
//
//	client := cache.NewCache(cache.NewMemoryStore(context.Background()))
//	locker := scheduler.NewCacheLocker(client, 10*time.Minute)
//	_, _ = locker.Lock(context.Background(), "job")
//
// Example: use the Redis cache driver
//
//	redisStore := rediscache.New(rediscache.Config{
//		Addr: "127.0.0.1:6379",
//	})
//	redisClient := cache.NewCache(redisStore)
//	redisLocker := scheduler.NewCacheLocker(redisClient, 10*time.Minute)
//	_, _ = redisLocker.Lock(context.Background(), "job")
func NewCacheLocker(client cache.LockAPI, ttl time.Duration) *CacheLocker {
	return &CacheLocker{client: client, ttl: ttl}
}

// Lock obtains a lock for the job name using the cache lock API.
// @group Locking
func (l *CacheLocker) Lock(ctx context.Context, key string) (gocron.Lock, error) {
	locked, err := l.client.TryLockCtx(ctx, l.lockKey(key), l.ttl)
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, errLockNotAcquired
	}
	return &cacheLock{
		client: l.client,
		key:    l.lockKey(key),
	}, nil
}

func (l *CacheLocker) lockKey(name string) string {
	return "gocron:lock:" + name
}

// RedisLocker is a simple gocron Locker backed by redis NX locks.
// It uses a single key per job and auto-expires after ttl.
type RedisLocker struct {
	client redisLockerClient
	ttl    time.Duration
}

// NewRedisLocker creates a RedisLocker with a client and TTL.
// The ttl is a lease duration: when it expires, another worker may acquire the
// same lock key. For long-running jobs, choose ttl >= worst-case runtime plus a
// safety buffer. If your runtime can exceed ttl, prefer a renewing/heartbeat lock strategy.
// @group Locking
//
// Example: create a redis-backed locker
//
//	client := redis.NewClient(&redis.Options{}) // replace with your client
//	locker := scheduler.NewRedisLocker(client, 10*time.Minute)
//	_, _ = locker.Lock(context.Background(), "job")
func NewRedisLocker(client redisLockerClient, ttl time.Duration) *RedisLocker {
	return &RedisLocker{client: client, ttl: ttl}
}

// Lock obtains a lock for the job name.
// @group Locking
//
// Example: acquire a lock
//
//	client := redis.NewClient(&redis.Options{})
//	locker := scheduler.NewRedisLocker(client, 10*time.Minute)
//	lock, _ := locker.Lock(context.Background(), "job")
//	_ = lock.Unlock(context.Background())
func (l *RedisLocker) Lock(ctx context.Context, key string) (gocron.Lock, error) {
	locked, err := l.client.SetNX(ctx, l.lockKey(key), "1", l.ttl).Result()
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, errLockNotAcquired
	}
	return &redisLock{
		client: l.client,
		key:    l.lockKey(key),
	}, nil
}

func (l *RedisLocker) lockKey(name string) string {
	return "gocron:lock:" + name
}

// Internal lock and error helpers.
type redisLock struct {
	client redisLockerClient
	key    string
}

type cacheLock struct {
	client cache.LockAPI
	key    string
}

func (l *cacheLock) Unlock(ctx context.Context) error {
	return l.client.UnlockCtx(ctx, l.key)
}

// Unlock releases the redis-backed lock.
// @group Locking
func (l *redisLock) Unlock(ctx context.Context) error {
	_, err := l.client.Del(ctx, l.key).Result()
	return err
}

var errLockNotAcquired = &lockError{"could not acquire lock"}

type lockError struct {
	msg string
}

func (e *lockError) Error() string { return e.msg }
