<p align="center">
  <img src="./docs/images/logo.png?v=2" width="400" alt="scheduler logo">
</p>

<p align="center">
    A fluent, Laravel-inspired scheduler for Go that wraps gocron with expressive APIs for defining, filtering, and controlling scheduled jobs.
</p>

<p align="center">
    <a href="https://pkg.go.dev/github.com/goforj/scheduler"><img src="https://pkg.go.dev/badge/github.com/goforj/scheduler.svg" alt="Go Reference"></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License: MIT"></a>
    <a href="https://github.com/goforj/scheduler/actions"><img src="https://github.com/goforj/scheduler/actions/workflows/test.yml/badge.svg" alt="Go Test"></a>
    <a href="https://golang.org"><img src="https://img.shields.io/badge/go-1.18+-blue?logo=go" alt="Go version"></a>
    <img src="https://img.shields.io/github/v/tag/goforj/scheduler?label=version&sort=semver" alt="Latest tag">
    <a href="https://codecov.io/gh/goforj/scheduler" ><img src="https://codecov.io/github/goforj/scheduler/graph/badge.svg?token=9KT46ZORP3"/></a>
<!-- test-count:embed:start -->
    <img src="https://img.shields.io/badge/test%20-count-183-brightgreen" alt="Tests">
<!-- test-count:embed:end -->
    <a href="https://goreportcard.com/report/github.com/goforj/scheduler"><img src="https://goreportcard.com/badge/github.com/goforj/scheduler" alt="Go Report Card"></a>
</p>

## Features

- Fluent, chainable API for intervals, cron strings, and calendar helpers (daily/weekly/monthly).
- Overlap protection with optional distributed locking plus per-job tags and metadata.
- Filters (weekdays/weekends/time windows) and hooks (before/after/success/failure) keep jobs predictable.
- Command execution helper for running CLI tasks with background mode and env-aware tagging.
- Auto-generated, compile-tested examples ensure docs and behavior stay in sync.

## Why scheduler?

Go has excellent low-level scheduling libraries, but defining real-world schedules often turns into a maze of cron strings, conditionals, and glue code.

`scheduler` provides a Laravel-style fluent API on top of gocron that lets you describe **when**, **how**, and **under what conditions** a job should run — without hiding what’s actually happening.

Everything remains explicit, testable, and inspectable, while staying pleasant to read and maintain.

## Example

```go
scheduler.NewJobBuilder(s).
    Name("reports:generate").
    Weekdays().
    Between("09:00", "17:00").
    WithoutOverlapping().
    DailyAt("10:30").
    Do(func() {
    generateReports()
})
```

## Runnable examples

Every function has a corresponding runnable example under [`./examples`](./examples).

These examples are **generated directly from the documentation blocks** of each function, ensuring the docs and code never drift. These are the same examples you see here in the README and GoDoc.

An automated test executes **every example** to verify it builds and runs successfully.

This guarantees all examples are valid, up-to-date, and remain functional as the API evolves.

<!-- api:embed:start -->

## API Index

| Group | Functions |
|------:|-----------|
| **Adapters** | [Lock](#lock) [Run](#run) [Unlock](#unlock) |
| **Commands** | [Command](#command) |
| **Concurrency** | [WithoutOverlapping](#withoutoverlapping) [WithoutOverlappingWithLocker](#withoutoverlappingwithlocker) |
| **Configuration** | [Timezone](#timezone) [WithCommandRunner](#withcommandrunner) [WithNowFunc](#withnowfunc) |
| **Construction** | [NewJobBuilder](#newjobbuilder) |
| **Diagnostics** | [CronExpr](#cronexpr) [Error](#error) [Job](#job) [PrintJobsList](#printjobslist) |
| **Execution** | [RunInBackground](#runinbackground) |
| **Filters** | [Between](#between) [Days](#days) [Environments](#environments) [Fridays](#fridays) [Mondays](#mondays) [Saturdays](#saturdays) [Skip](#skip) [Sundays](#sundays) [Thursdays](#thursdays) [Tuesdays](#tuesdays) [UnlessBetween](#unlessbetween) [Wednesdays](#wednesdays) [Weekdays](#weekdays) [Weekends](#weekends) [When](#when) |
| **Hooks** | [After](#after) [Before](#before) [OnFailure](#onfailure) [OnSuccess](#onsuccess) |
| **Locking** | [NewRedisLocker](#newredislocker) |
| **Metadata** | [JobMetadata](#jobmetadata) [Name](#name) |
| **Scheduling** | [Cron](#cron) [Daily](#daily) [DailyAt](#dailyat) [DaysOfMonth](#daysofmonth) [Do](#do) [Every](#every) [EveryFifteenMinutes](#everyfifteenminutes) [EveryFifteenSeconds](#everyfifteenseconds) [EveryFiveMinutes](#everyfiveminutes) [EveryFiveSeconds](#everyfiveseconds) [EveryFourHours](#everyfourhours) [EveryFourMinutes](#everyfourminutes) [EveryMinute](#everyminute) [EveryOddHour](#everyoddhour) [EverySecond](#everysecond) [EverySixHours](#everysixhours) [EveryTenMinutes](#everytenminutes) [EveryTenSeconds](#everytenseconds) [EveryThirtyMinutes](#everythirtyminutes) [EveryThirtySeconds](#everythirtyseconds) [EveryThreeHours](#everythreehours) [EveryThreeMinutes](#everythreeminutes) [EveryTwentySeconds](#everytwentyseconds) [EveryTwoHours](#everytwohours) [EveryTwoMinutes](#everytwominutes) [EveryTwoSeconds](#everytwoseconds) [Hourly](#hourly) [HourlyAt](#hourlyat) [Hours](#hours) [LastDayOfMonth](#lastdayofmonth) [Minutes](#minutes) [Monthly](#monthly) [MonthlyOn](#monthlyon) [Quarterly](#quarterly) [QuarterlyOn](#quarterlyon) [Seconds](#seconds) [TwiceDaily](#twicedaily) [TwiceDailyAt](#twicedailyat) [TwiceMonthly](#twicemonthly) [Weekly](#weekly) [WeeklyOn](#weeklyon) [Yearly](#yearly) [YearlyOn](#yearlyon) |
| **State management** | [RetainState](#retainstate) |


## Adapters

### <a id="lock"></a>Lock

Lock invokes the underlying function.

_Example: acquire a lock_

```go
client := redis.NewClient(&redis.Options{})
locker := scheduler.NewRedisLocker(client, time.Minute)
lock, _ := locker.Lock(context.Background(), "job")
_ = lock.Unlock(context.Background())
```

### <a id="run"></a>Run

Run executes the underlying function.

_Example: execute the wrapped function_

```go
runner := scheduler.CommandRunnerFunc(func(ctx context.Context, exe string, args []string) error {
	return nil
})
_ = runner.Run(context.Background(), "echo", []string{"hi"})
```

### <a id="unlock"></a>Unlock

Unlock invokes the underlying function.

## Commands

### <a id="command"></a>Command

Command executes the current binary with the given subcommand and variadic args.

_Example: run a CLI subcommand on schedule_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).
	Cron("0 0 * * *").
	Command("jobs:purge", "--force")
```

## Concurrency

### <a id="withoutoverlapping"></a>WithoutOverlapping

WithoutOverlapping ensures the job does not run concurrently.

_Example: prevent overlapping runs of a slow task_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).
	WithoutOverlapping().
	EveryFiveSeconds().
	Do(func() { time.Sleep(7 * time.Second) })
```

### <a id="withoutoverlappingwithlocker"></a>WithoutOverlappingWithLocker

WithoutOverlappingWithLocker ensures the job does not run concurrently across distributed systems using the provided locker.

_Example: use a distributed locker_

```go
locker := scheduler.LockerFunc(func(ctx context.Context, key string) (gocron.Lock, error) {
	return scheduler.LockFunc(func(context.Context) error { return nil }), nil
})

s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).
	WithoutOverlappingWithLocker(locker).
	EveryMinute().
	Do(func() {})
```

## Configuration

### <a id="timezone"></a>Timezone

Timezone sets a timezone string for the job (not currently applied to gocron Scheduler).

_Example: tag jobs with a timezone_

```go
scheduler.NewJobBuilder(nil).
	Timezone("America/New_York").
	Daily()
```

### <a id="withcommandrunner"></a>WithCommandRunner

WithCommandRunner overrides command execution (default: exec.CommandContext).

_Example: swap in a custom runner_

```go
runner := scheduler.CommandRunnerFunc(func(_ context.Context, exe string, args []string) error {
	fmt.Println(exe, args)
	return nil
})

builder := scheduler.NewJobBuilder(nil).
	WithCommandRunner(runner)
fmt.Printf("%T\n", builder)
```

### <a id="withnowfunc"></a>WithNowFunc

WithNowFunc overrides current time (default: time.Now). Useful for tests.

_Example: freeze time for predicates_

```go
fixed := func() time.Time { return time.Unix(0, 0) }
scheduler.NewJobBuilder(nil).WithNowFunc(fixed)
```

## Construction

### <a id="newjobbuilder"></a>NewJobBuilder

NewJobBuilder creates a new JobBuilder with the provided scheduler.

_Example: create a builder and schedule a heartbeat_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()
scheduler.NewJobBuilder(s).EverySecond().Do(func() {})
```

## Diagnostics

### <a id="cronexpr"></a>CronExpr

CronExpr returns the cron expression string configured for this job.

_Example: inspect the stored cron expression_

```go
builder := scheduler.NewJobBuilder(nil).Cron("0 9 * * *")
fmt.Println(builder.CronExpr())
```

### <a id="error"></a>Error

Error returns the error if any occurred during job scheduling.

_Example: validate a malformed schedule_

```go
builder := scheduler.NewJobBuilder(nil).DailyAt("bad")
fmt.Println(builder.Error())
```

### <a id="job"></a>Job

Job returns the last scheduled gocron.Job instance, if available.

_Example: capture the last job handle_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

b := scheduler.NewJobBuilder(s).EverySecond().Do(func() {})
fmt.Println(b.Job() != nil)
```

### <a id="printjobslist"></a>PrintJobsList

PrintJobsList renders and prints the scheduler job table to stdout.

_Example: print current jobs_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).
	EverySecond().
	Name("heartbeat").
	Do(func() {})

scheduler.NewJobBuilder(s).PrintJobsList()
```

## Execution

### <a id="runinbackground"></a>RunInBackground

RunInBackground runs command/exec tasks in a goroutine.

_Example: allow command jobs to run async_

```go
scheduler.NewJobBuilder(nil).
	RunInBackground().
	Command("noop")
```

## Filters

### <a id="between"></a>Between

Between limits the job to run between the provided HH:MM times (inclusive).

_Example: allow execution during business hours_

```go
scheduler.NewJobBuilder(nil).
	Between("09:00", "17:00").
	EveryMinute()
```

### <a id="days"></a>Days

Days limits the job to a specific set of weekdays.

_Example: pick custom weekdays_

```go
scheduler.NewJobBuilder(nil).
	Days(time.Monday, time.Wednesday, time.Friday).
	DailyAt("07:00")
```

### <a id="environments"></a>Environments

Environments restricts job registration to specific environment names (e.g. "production", "staging").

_Example: only register in production_

```go
scheduler.NewJobBuilder(nil).Environments("production").Daily()
```

### <a id="fridays"></a>Fridays

Fridays limits the job to Fridays.

### <a id="mondays"></a>Mondays

Mondays limits the job to Mondays.

### <a id="saturdays"></a>Saturdays

Saturdays limits the job to Saturdays.

### <a id="skip"></a>Skip

Skip prevents scheduling the job if the provided condition returns true.

_Example: suppress jobs based on a switch_

```go
enabled := false
scheduler.NewJobBuilder(nil).
	Skip(func() bool { return !enabled }).
	Daily()
```

### <a id="sundays"></a>Sundays

Sundays limits the job to Sundays.

### <a id="thursdays"></a>Thursdays

Thursdays limits the job to Thursdays.

### <a id="tuesdays"></a>Tuesdays

Tuesdays limits the job to Tuesdays.

### <a id="unlessbetween"></a>UnlessBetween

UnlessBetween prevents the job from running between the provided HH:MM times.

_Example: pause execution overnight_

```go
scheduler.NewJobBuilder(nil).
	UnlessBetween("22:00", "06:00").
	EveryMinute()
```

### <a id="wednesdays"></a>Wednesdays

Wednesdays limits the job to Wednesdays.

### <a id="weekdays"></a>Weekdays

Weekdays limits the job to run only on weekdays (Mon-Fri).

_Example: weekday-only execution_

```go
scheduler.NewJobBuilder(nil).Weekdays().DailyAt("09:00")
```

### <a id="weekends"></a>Weekends

Weekends limits the job to run only on weekends (Sat-Sun).

_Example: weekend-only execution_

```go
scheduler.NewJobBuilder(nil).Weekends().DailyAt("10:00")
```

### <a id="when"></a>When

When only schedules the job if the provided condition returns true.

_Example: guard scheduling with a flag_

```go
flag := true
scheduler.NewJobBuilder(nil).
	When(func() bool { return flag }).
	Daily()
```

## Hooks

### <a id="after"></a>After

After sets a hook to run after task execution.

_Example: add an after hook_

```go
scheduler.NewJobBuilder(nil).
	After(func() { println("after") }).
	Daily()
```

### <a id="before"></a>Before

Before sets a hook to run before task execution.

_Example: add a before hook_

```go
scheduler.NewJobBuilder(nil).
	Before(func() { println("before") }).
	Daily()
```

### <a id="onfailure"></a>OnFailure

OnFailure sets a hook to run after failed task execution.

_Example: record failures_

```go
scheduler.NewJobBuilder(nil).
	OnFailure(func() { println("failure") }).
	Daily()
```

### <a id="onsuccess"></a>OnSuccess

OnSuccess sets a hook to run after successful task execution.

_Example: record success_

```go
scheduler.NewJobBuilder(nil).
	OnSuccess(func() { println("success") }).
	Daily()
```

## Locking

### <a id="newredislocker"></a>NewRedisLocker

NewRedisLocker creates a RedisLocker with a client and TTL.

_Example: create a redis-backed locker_

```go
client := redis.NewClient(&redis.Options{}) // replace with your client
locker := scheduler.NewRedisLocker(client, time.Minute)
_, _ = locker.Lock(context.Background(), "job")
```

## Metadata

### <a id="jobmetadata"></a>JobMetadata

JobMetadata returns a copy of the tracked job metadata keyed by job ID.

_Example: inspect scheduled jobs_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

b := scheduler.NewJobBuilder(s).EverySecond().Do(func() {})
for id, meta := range b.JobMetadata() {
	_ = id
	_ = meta.Name
}
```

### <a id="name"></a>Name

Name sets an explicit job name.

_Example: label a job for logging_

```go
scheduler.NewJobBuilder(nil).
	Name("cache:refresh").
	HourlyAt(15)
```

## Scheduling

### <a id="cron"></a>Cron

Cron sets the cron expression for the job.

_Example: configure a cron expression_

```go
builder := scheduler.NewJobBuilder(nil).Cron("15 3 * * *")
fmt.Println(builder.CronExpr())
```

### <a id="daily"></a>Daily

Daily schedules the job to run once per day at midnight.

_Example: nightly task_

```go
scheduler.NewJobBuilder(nil).Daily()
```

### <a id="dailyat"></a>DailyAt

DailyAt schedules the job to run daily at a specific time (e.g., "13:00").

_Example: run at lunch time daily_

```go
scheduler.NewJobBuilder(nil).DailyAt("12:30")
```

### <a id="daysofmonth"></a>DaysOfMonth

DaysOfMonth schedules the job to run on specific days of the month at a given time.

_Example: run on the 5th and 20th of each month_

```go
scheduler.NewJobBuilder(nil).DaysOfMonth([]int{5, 20}, "07:15")
```

### <a id="do"></a>Do

Do schedules the job with the provided task function.

_Example: create a named cron job_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).
Name("cleanup").
Cron("0 0 * * *").
Do(func() {})
```

### <a id="every"></a>Every

Every schedules a job to run every X seconds, minutes, or hours.

_Example: fluently choose an interval_

```go
scheduler.NewJobBuilder(nil).
	Every(10).
	Minutes()
```

### <a id="everyfifteenminutes"></a>EveryFifteenMinutes

EveryFifteenMinutes schedules the job to run every 15 minutes.

_Example: run every fifteen minutes_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).EveryFifteenMinutes().Do(func() {})
```

### <a id="everyfifteenseconds"></a>EveryFifteenSeconds

EveryFifteenSeconds schedules the job to run every 15 seconds.

_Example: run at fifteen-second cadence_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).EveryFifteenSeconds().Do(func() {})
```

### <a id="everyfiveminutes"></a>EveryFiveMinutes

EveryFiveMinutes schedules the job to run every 5 minutes.

_Example: run every five minutes_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).EveryFiveMinutes().Do(func() {})
```

### <a id="everyfiveseconds"></a>EveryFiveSeconds

EveryFiveSeconds schedules the job to run every 5 seconds.

_Example: space out work every five seconds_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).EveryFiveSeconds().Do(func() {})
```

### <a id="everyfourhours"></a>EveryFourHours

EveryFourHours schedules the job to run every four hours at the specified minute.

_Example: run every four hours_

```go
scheduler.NewJobBuilder(nil).EveryFourHours(25)
```

### <a id="everyfourminutes"></a>EveryFourMinutes

EveryFourMinutes schedules the job to run every 4 minutes.

_Example: run every four minutes_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).EveryFourMinutes().Do(func() {})
```

### <a id="everyminute"></a>EveryMinute

EveryMinute schedules the job to run every 1 minute.

_Example: run a task each minute_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).EveryMinute().Do(func() {})
```

### <a id="everyoddhour"></a>EveryOddHour

EveryOddHour schedules the job to run every odd-numbered hour at the specified minute.

_Example: run every odd hour_

```go
scheduler.NewJobBuilder(nil).EveryOddHour(10)
```

### <a id="everysecond"></a>EverySecond

EverySecond schedules the job to run every 1 second.

_Example: heartbeat job each second_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).EverySecond().Do(func() {})
```

### <a id="everysixhours"></a>EverySixHours

EverySixHours schedules the job to run every six hours at the specified minute.

_Example: run every six hours_

```go
scheduler.NewJobBuilder(nil).EverySixHours(30)
```

### <a id="everytenminutes"></a>EveryTenMinutes

EveryTenMinutes schedules the job to run every 10 minutes.

_Example: run every ten minutes_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).EveryTenMinutes().Do(func() {})
```

### <a id="everytenseconds"></a>EveryTenSeconds

EveryTenSeconds schedules the job to run every 10 seconds.

_Example: poll every ten seconds_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).EveryTenSeconds().Do(func() {})
```

### <a id="everythirtyminutes"></a>EveryThirtyMinutes

EveryThirtyMinutes schedules the job to run every 30 minutes.

_Example: run every thirty minutes_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).EveryThirtyMinutes().Do(func() {})
```

### <a id="everythirtyseconds"></a>EveryThirtySeconds

EveryThirtySeconds schedules the job to run every 30 seconds.

_Example: execute every thirty seconds_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).EveryThirtySeconds().Do(func() {})
```

### <a id="everythreehours"></a>EveryThreeHours

EveryThreeHours schedules the job to run every three hours at the specified minute.

_Example: run every three hours_

```go
scheduler.NewJobBuilder(nil).EveryThreeHours(20)
```

### <a id="everythreeminutes"></a>EveryThreeMinutes

EveryThreeMinutes schedules the job to run every 3 minutes.

_Example: run every three minutes_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).EveryThreeMinutes().Do(func() {})
```

### <a id="everytwentyseconds"></a>EveryTwentySeconds

EveryTwentySeconds schedules the job to run every 20 seconds.

_Example: run once every twenty seconds_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).EveryTwentySeconds().Do(func() {})
```

### <a id="everytwohours"></a>EveryTwoHours

EveryTwoHours schedules the job to run every two hours at the specified minute.

_Example: run every two hours_

```go
scheduler.NewJobBuilder(nil).EveryTwoHours(15)
```

### <a id="everytwominutes"></a>EveryTwoMinutes

EveryTwoMinutes schedules the job to run every 2 minutes.

_Example: job that runs every two minutes_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).EveryTwoMinutes().Do(func() {})
```

### <a id="everytwoseconds"></a>EveryTwoSeconds

EveryTwoSeconds schedules the job to run every 2 seconds.

_Example: throttle a task to two seconds_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).EveryTwoSeconds().Do(func() {})
```

### <a id="hourly"></a>Hourly

Hourly schedules the job to run every hour.

_Example: run something hourly_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).Hourly().Do(func() {})
```

### <a id="hourlyat"></a>HourlyAt

HourlyAt schedules the job to run every hour at the specified minute.

_Example: run at the 5th minute of each hour_

```go
scheduler.NewJobBuilder(nil).HourlyAt(5)
```

### <a id="hours"></a>Hours

Hours schedules the job to run every X hours.

_Example: build an hourly cadence_

```go
scheduler.NewJobBuilder(nil).Every(6).Hours()
```

### <a id="lastdayofmonth"></a>LastDayOfMonth

LastDayOfMonth schedules the job to run on the last day of each month at a specific time.

_Example: run on the last day of the month_

```go
scheduler.NewJobBuilder(nil).LastDayOfMonth("23:30")
```

### <a id="minutes"></a>Minutes

Minutes schedules the job to run every X minutes.

_Example: chain a minute-based interval_

```go
scheduler.NewJobBuilder(nil).Every(15).Minutes()
```

### <a id="monthly"></a>Monthly

Monthly schedules the job to run on the first day of each month at midnight.

_Example: first-of-month billing_

```go
scheduler.NewJobBuilder(nil).Monthly()
```

### <a id="monthlyon"></a>MonthlyOn

MonthlyOn schedules the job to run on a specific day of the month at a given time.

_Example: run on the 15th of each month_

```go
scheduler.NewJobBuilder(nil).MonthlyOn(15, "09:30")
```

### <a id="quarterly"></a>Quarterly

Quarterly schedules the job to run on the first day of each quarter at midnight.

_Example: quarterly trigger_

```go
scheduler.NewJobBuilder(nil).Quarterly()
```

### <a id="quarterlyon"></a>QuarterlyOn

QuarterlyOn schedules the job to run on a specific day of each quarter at a given time.

_Example: quarterly on a specific day_

```go
scheduler.NewJobBuilder(nil).QuarterlyOn(3, "12:00")
```

### <a id="seconds"></a>Seconds

Seconds schedules the job to run every X seconds.

_Example: run a task every few seconds_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

scheduler.NewJobBuilder(s).
	Every(3).
	Seconds().
	Do(func() {})
```

### <a id="twicedaily"></a>TwiceDaily

TwiceDaily schedules the job to run daily at two specified hours (e.g., 1 and 13).

_Example: run two times per day_

```go
scheduler.NewJobBuilder(nil).TwiceDaily(1, 13)
```

### <a id="twicedailyat"></a>TwiceDailyAt

TwiceDailyAt schedules the job to run daily at two specified times (e.g., 1:15 and 13:15).

_Example: run twice daily at explicit minutes_

```go
scheduler.NewJobBuilder(nil).TwiceDailyAt(1, 13, 15)
```

### <a id="twicemonthly"></a>TwiceMonthly

TwiceMonthly schedules the job to run on two specific days of the month at the given time.

_Example: run on two days each month_

```go
scheduler.NewJobBuilder(nil).TwiceMonthly(1, 15, "10:00")
```

### <a id="weekly"></a>Weekly

Weekly schedules the job to run once per week on Sunday at midnight.

_Example: weekly maintenance_

```go
scheduler.NewJobBuilder(nil).Weekly()
```

### <a id="weeklyon"></a>WeeklyOn

WeeklyOn schedules the job to run weekly on a specific day of the week and time.
Day uses 0 = Sunday through 6 = Saturday.

_Example: run each Monday at 08:00_

```go
scheduler.NewJobBuilder(nil).WeeklyOn(1, "8:00")
```

### <a id="yearly"></a>Yearly

Yearly schedules the job to run on January 1st every year at midnight.

_Example: yearly trigger_

```go
scheduler.NewJobBuilder(nil).Yearly()
```

### <a id="yearlyon"></a>YearlyOn

YearlyOn schedules the job to run every year on a specific month, day, and time.

_Example: yearly on a specific date_

```go
scheduler.NewJobBuilder(nil).YearlyOn(12, 25, "06:45")
```

## State management

### <a id="retainstate"></a>RetainState

RetainState allows the job to retain its state after execution.

_Example: reuse interval configuration for multiple jobs_

```go
s, _ := gocron.NewScheduler()
s.Start()
defer s.Shutdown()

builder := scheduler.NewJobBuilder(s).EverySecond().RetainState()
builder.Do(func() {})
builder.Do(func() {})
```
<!-- api:embed:end -->
