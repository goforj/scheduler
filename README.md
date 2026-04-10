<p align="center">
  <img src="./docs/images/logo.png?v=2" width="300" alt="scheduler logo">
</p>

<p align="center">
    A fluent, Laravel-inspired scheduler for Go that wraps gocron with expressive APIs for defining, filtering, and controlling scheduled jobs.
</p>

<p align="center">
    <a href="https://pkg.go.dev/github.com/goforj/scheduler/v2"><img src="https://pkg.go.dev/badge/github.com/goforj/scheduler/v2.svg" alt="Go Reference"></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License: MIT"></a>
    <a href="https://github.com/goforj/scheduler/actions"><img src="https://github.com/goforj/scheduler/actions/workflows/test.yml/badge.svg" alt="Go Test"></a>
    <a href="https://golang.org"><img src="https://img.shields.io/badge/go-1.18+-blue?logo=go" alt="Go version"></a>
    <img src="https://img.shields.io/github/v/tag/goforj/scheduler?label=version&sort=semver" alt="Latest tag">
    <a href="https://codecov.io/gh/goforj/scheduler" ><img src="https://codecov.io/github/goforj/scheduler/graph/badge.svg?token=9KT46ZORP3"/></a>
<!-- test-count:embed:start -->
    <img src="https://img.shields.io/badge/tests-219-brightgreen" alt="Tests">
<!-- test-count:embed:end -->
    <a href="https://goreportcard.com/report/github.com/goforj/scheduler/v2"><img src="https://goreportcard.com/badge/github.com/goforj/scheduler/v2" alt="Go Report Card"></a>
</p>

## Features

- Fluent, chainable API for intervals, cron strings, and calendar helpers (daily/weekly/monthly).
- Overlap protection with optional distributed locking plus per-job tags and metadata.
- Filters (weekdays/weekends/time windows) and hooks (before/after/success/failure) keep jobs predictable.
- Command execution helper for running CLI tasks with background mode and env-aware tagging.
- Auto-generated, compile-tested examples ensure docs and behavior stay in sync.

## Why scheduler?

Go has excellent low-level scheduling libraries, but defining real-world schedules often turns into a maze of cron strings, conditionals, and glue code.

`scheduler` provides a Laravel-style fluent API on top of gocron that lets you describe **when**, **how**, and **under what conditions** a job should run - without hiding what’s actually happening.

Everything remains explicit, testable, and inspectable, while staying pleasant to read and maintain.

## Installation

```bash
go get github.com/goforj/scheduler/v2
```

If you use distributed locking, there are two paths:

- Bring your own `gocron.Locker` implementation and install whatever backend client it needs.
- Use `NewCacheLocker` with `github.com/goforj/cache` and install only the cache driver you plan to use.

For example, with a cache-backed Redis locker:

```bash
go get github.com/goforj/cache
go get github.com/goforj/cache/driver/rediscache
go get github.com/redis/go-redis/v9
```

## Quick Start

### Basic

```go
s := scheduler.New()
defer s.Stop()

s.EveryMinute().Name("cleanup").Do(func(context.Context) error { return runCleanup() }) // run in-process cleanup every minute
s.DailyAt("09:00").Weekdays().Name("reports:morning").Do(func(context.Context) error { return sendMorningReport() }) // weekdays at 09:00
s.Cron("0 0 * * *").Command("reports:purge", "--force") // run app subcommand nightly
s.Cron("*/15 * * * *").Exec("/usr/bin/env", "echo", "heartbeat") // run external executable every 15 minutes
s.EveryFiveMinutes().WithoutOverlapping().Name("sync:inventory").Do(func(context.Context) error { return syncInventory() }) // prevent overlapping runs
s.Cron("0 * * * *").When(func() bool { return isPrimaryNode() }).Name("rebalance").Do(func(context.Context) error { return rebalance() }) // run only when condition passes
```

### Advanced (kitchen sink)

```go
s := scheduler.New()
defer s.Stop()

s.
	Name("reports:generate").
	Timezone("America/New_York").
	Weekdays().
	Between("09:00", "17:00").
	WithoutOverlapping().
	Before(func(context.Context) { markJobStart("reports:generate") }).
	OnSuccess(func(context.Context) { notifySuccess("reports:generate") }).
	OnFailure(func(_ context.Context, err error) { notifyFailure("reports:generate", err) }).
	DailyAt("10:30").
	Do(func(context.Context) error { return generateReports() })

s.
	Name("reconcile:daily").
	RunInBackground().
	Cron("0 3 * * *").
	Command("billing:reconcile", "--retry=3")
```

## List jobs as an ASCII table

```go
package main

import (
	"context"
	"github.com/goforj/scheduler/v2"
)

func main() {
	s := scheduler.New()
	defer s.Stop()

	s.EveryMinute().Name("cleanup").Do(func(context.Context) error { return nil }) // run cleanup every minute
	s.DailyAt("10:30").Weekdays().Name("reports:generate").Do(func(context.Context) error { return nil }) // run reports on weekdays at 10:30
	s.Cron("0 0 * * *").Command("reports:purge", "--force") // run app subcommand nightly at midnight

	s.PrintJobsList()
}
```

Example output:

```
+------------------------------------------------------------------------------------------------------------------------+
| Scheduler Jobs › (3)                                                                                                  |
+------------------+----------+----------------+-----------------------+----------------------+--------------------------+
| Name             | Type     | Schedule       | Handler               | Next Run             | Tags                     |
+------------------+----------+----------------+-----------------------+----------------------+--------------------------+
| cleanup          | function | every 1m       | main.main (anon func) | in 1m Mar 3 2:16AM  | env=local                |
| reports:generate | function | cron 30 10 * * * | main.main (anon func) | in 8h Mar 3 10:30AM | env=local                |
| reports:purge    | command  | cron 0 0 * * * | -                     | in 21h Mar 4 12:00AM | env=local, args="--force" |
+------------------+----------+----------------+-----------------------+----------------------+--------------------------+
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
| **Calendar** | [Daily](#daily) [DailyAt](#dailyat) [DaysOfMonth](#daysofmonth) [LastDayOfMonth](#lastdayofmonth) [Monthly](#monthly) [MonthlyOn](#monthlyon) [Quarterly](#quarterly) [QuarterlyOn](#quarterlyon) [TwiceDaily](#twicedaily) [TwiceDailyAt](#twicedailyat) [TwiceMonthly](#twicemonthly) [Weekly](#weekly) [WeeklyOn](#weeklyon) [Yearly](#yearly) [YearlyOn](#yearlyon) |
| **Commands** | [Command](#command) [Exec](#exec) |
| **Concurrency** | [WithoutOverlapping](#withoutoverlapping) [WithoutOverlappingWithLocker](#withoutoverlappingwithlocker) |
| **Configuration** | [Timezone](#timezone) [WithCommandRunner](#withcommandrunner) [WithNowFunc](#withnowfunc) |
| **Construction** | [New](#new) [NewWithError](#newwitherror) |
| **Diagnostics** | [CronExpr](#cronexpr) [Error](#error) [Job](#job) [Jobs](#jobs) [PrintJobsList](#printjobslist) |
| **Execution** | [RunInBackground](#runinbackground) |
| **Filters** | [Between](#between) [Days](#days) [Environments](#environments) [Fridays](#fridays) [Mondays](#mondays) [Saturdays](#saturdays) [Skip](#skip) [Sundays](#sundays) [Thursdays](#thursdays) [Tuesdays](#tuesdays) [UnlessBetween](#unlessbetween) [Wednesdays](#wednesdays) [Weekdays](#weekdays) [Weekends](#weekends) [When](#when) |
| **Hooks** | [After](#after) [Before](#before) [OnFailure](#onfailure) [OnSuccess](#onsuccess) |
| **Interop** | [GocronScheduler](#gocronscheduler) |
| **Intervals** | [Every](#every) [EveryDuration](#everyduration) [EveryFifteenMinutes](#everyfifteenminutes) [EveryFifteenSeconds](#everyfifteenseconds) [EveryFiveMinutes](#everyfiveminutes) [EveryFiveSeconds](#everyfiveseconds) [EveryFourHours](#everyfourhours) [EveryFourMinutes](#everyfourminutes) [EveryMinute](#everyminute) [EveryOddHour](#everyoddhour) [EverySecond](#everysecond) [EverySixHours](#everysixhours) [EveryTenMinutes](#everytenminutes) [EveryTenSeconds](#everytenseconds) [EveryThirtyMinutes](#everythirtyminutes) [EveryThirtySeconds](#everythirtyseconds) [EveryThreeHours](#everythreehours) [EveryThreeMinutes](#everythreeminutes) [EveryTwentySeconds](#everytwentyseconds) [EveryTwoHours](#everytwohours) [EveryTwoMinutes](#everytwominutes) [EveryTwoSeconds](#everytwoseconds) [Hourly](#hourly) [HourlyAt](#hourlyat) [Hours](#hours) [Minutes](#minutes) [Seconds](#seconds) |
| **Lifecycle** | [Shutdown](#shutdown) [Start](#start) [Stop](#stop) |
| **Locking** | [NewCacheLocker](#newcachelocker) [NewRedisLocker](#newredislocker) |
| **Metadata** | [JobMetadata](#jobmetadata) [JobsInfo](#jobsinfo) [Name](#name) |
| **Runtime control** | [IsJobPaused](#isjobpaused) [IsPausedAll](#ispausedall) [Observe](#observe) [PauseAll](#pauseall) [PauseJob](#pausejob) [ResumeAll](#resumeall) [ResumeJob](#resumejob) |
| **State management** | [RetainState](#retainstate) |
| **Triggers** | [Cron](#cron) [Do](#do) |


## Adapters

### <a id="lock"></a>Lock

Lock invokes the underlying function.

```go
client := redis.NewClient(&redis.Options{})
locker := scheduler.NewRedisLocker(client, 10*time.Minute)
lock, _ := locker.Lock(context.Background(), "job")
_ = lock.Unlock(context.Background())
```

### <a id="run"></a>Run

Run executes the underlying function.

```go
runner := scheduler.CommandRunnerFunc(func(ctx context.Context, exe string, args []string) error {
	return nil
})
_ = runner.Run(context.Background(), "echo", []string{"hi"})
```

### <a id="unlock"></a>Unlock

Unlock invokes the underlying function.

```go
lock := scheduler.LockFunc(func(context.Context) error { return nil })
_ = lock.Unlock(context.Background())
```

## Calendar

### <a id="daily"></a>Daily

Daily schedules the job to run once per day at midnight.

```go
scheduler.New().Daily()
```

### <a id="dailyat"></a>DailyAt

DailyAt schedules the job to run daily at a specific time (e.g., "13:00").

```go
scheduler.New().DailyAt("12:30")
```

### <a id="daysofmonth"></a>DaysOfMonth

DaysOfMonth schedules the job to run on specific days of the month at a given time.

```go
scheduler.New().DaysOfMonth([]int{5, 20}, "07:15")
```

### <a id="lastdayofmonth"></a>LastDayOfMonth

LastDayOfMonth schedules the job to run on the last day of each month at a specific time.

```go
scheduler.New().LastDayOfMonth("23:30")
```

### <a id="monthly"></a>Monthly

Monthly schedules the job to run on the first day of each month at midnight.

```go
scheduler.New().Monthly()
```

### <a id="monthlyon"></a>MonthlyOn

MonthlyOn schedules the job to run on a specific day of the month at a given time.

```go
scheduler.New().MonthlyOn(15, "09:30")
```

### <a id="quarterly"></a>Quarterly

Quarterly schedules the job to run on the first day of each quarter at midnight.

```go
scheduler.New().Quarterly()
```

### <a id="quarterlyon"></a>QuarterlyOn

QuarterlyOn schedules the job to run on a specific day of each quarter at a given time.

```go
scheduler.New().QuarterlyOn(3, "12:00")
```

### <a id="twicedaily"></a>TwiceDaily

TwiceDaily schedules the job to run daily at two specified hours (e.g., 1 and 13).

```go
scheduler.New().TwiceDaily(1, 13)
```

### <a id="twicedailyat"></a>TwiceDailyAt

TwiceDailyAt schedules the job to run daily at two specified times (e.g., 1:15 and 13:15).

```go
scheduler.New().TwiceDailyAt(1, 13, 15)
```

### <a id="twicemonthly"></a>TwiceMonthly

TwiceMonthly schedules the job to run on two specific days of the month at the given time.

```go
scheduler.New().TwiceMonthly(1, 15, "10:00")
```

### <a id="weekly"></a>Weekly

Weekly schedules the job to run once per week on Sunday at midnight.

```go
scheduler.New().Weekly()
```

### <a id="weeklyon"></a>WeeklyOn

WeeklyOn schedules the job to run weekly on a specific day of the week and time.
Day uses 0 = Sunday through 6 = Saturday.

```go
scheduler.New().WeeklyOn(1, "8:00")
```

### <a id="yearly"></a>Yearly

Yearly schedules the job to run on January 1st every year at midnight.

```go
scheduler.New().Yearly()
```

### <a id="yearlyon"></a>YearlyOn

YearlyOn schedules the job to run every year on a specific month, day, and time.

```go
scheduler.New().YearlyOn(12, 25, "06:45")
```

## Commands

### <a id="command"></a>Command

Command executes the current binary with the given subcommand and variadic args.
It does not run arbitrary system executables; use Exec for that.

```go
scheduler.New().Cron("0 0 * * *").Command("jobs:purge", "--force")
```

### <a id="exec"></a>Exec

Exec runs an external executable with variadic args.

```go
scheduler.New().Cron("0 0 * * *").Exec("/usr/bin/env", "echo", "hello")
```

## Concurrency

### <a id="withoutoverlapping"></a>WithoutOverlapping

WithoutOverlapping ensures the job does not run concurrently.

```go
scheduler.New().
	WithoutOverlapping().
	EveryFiveSeconds().
	Do(func(context.Context) error { time.Sleep(7 * time.Second); return nil })
```

### <a id="withoutoverlappingwithlocker"></a>WithoutOverlappingWithLocker

WithoutOverlappingWithLocker ensures the job does not run concurrently across distributed systems using the provided locker.

```go
locker := scheduler.LockerFunc(func(ctx context.Context, key string) (gocron.Lock, error) {
	return scheduler.LockFunc(func(context.Context) error { return nil }), nil
})

scheduler.New().
	WithoutOverlappingWithLocker(locker).
	EveryMinute().
	Do(func(context.Context) error { return nil })
```

## Configuration

### <a id="timezone"></a>Timezone

Timezone sets a timezone string for the job (not currently applied to gocron Scheduler).

```go
scheduler.New().Timezone("America/New_York").Daily()
```

### <a id="withcommandrunner"></a>WithCommandRunner

WithCommandRunner overrides command execution (default: exec.CommandContext).

```go
runner := scheduler.CommandRunnerFunc(func(_ context.Context, exe string, args []string) error {
	_ = exe
	_ = args
	return nil
})

builder := scheduler.New().WithCommandRunner(runner)
_ = builder
```

### <a id="withnowfunc"></a>WithNowFunc

WithNowFunc overrides current time (default: time.Now). Useful for tests.

```go
fixed := func() time.Time { return time.Unix(0, 0) }
scheduler.New().WithNowFunc(fixed)
```

## Construction

### <a id="new"></a>New

New creates and starts a scheduler facade.
It panics only if gocron scheduler construction fails.

```go
s := scheduler.New()
defer s.Stop()
s.Every(15).Seconds().Do(func(context.Context) error { return nil })
```

### <a id="newwitherror"></a>NewWithError

NewWithError creates and starts a scheduler facade and returns setup errors.

```go
s, err := scheduler.NewWithError()
if err != nil {
	panic(err)
}
defer s.Stop()
```

## Diagnostics

### <a id="cronexpr"></a>CronExpr

CronExpr returns the cron expression string configured for this job.

```go
builder := scheduler.New().Cron("0 9 * * *")
fmt.Println(builder.CronExpr())
// Output: 0 9 * * *
```

### <a id="error"></a>Error

Error returns the error if any occurred during job scheduling.

```go
builder := scheduler.New().DailyAt("bad")
fmt.Println(builder.Error())
// Output: invalid DailyAt time format: invalid time format (expected HH:MM): "bad"
```

### <a id="job"></a>Job

Job returns the last scheduled gocron.Job instance, if available.

```go
b := scheduler.New().EverySecond().Do(func(context.Context) error { return nil })
fmt.Println(b.Job() != nil)
// Output: true
```

### <a id="jobs"></a>Jobs

Jobs returns scheduled jobs from the underlying scheduler.

### <a id="printjobslist"></a>PrintJobsList

PrintJobsList renders and prints the scheduler job table to stdout.

```go
s := scheduler.New()
defer s.Stop()
s.EverySecond().Name("heartbeat").Do(func(context.Context) error { return nil })
s.PrintJobsList()
// Output:
// +------------------------------------------------------------------------------------------+
// | Scheduler Jobs › (1)
// +-----------+----------+----------+-----------------------+--------------------+-----------+
// | Name      | Type     | Schedule | Handler               | Next Run           | Tags      |
// +-----------+----------+----------+-----------------------+--------------------+-----------+
// | heartbeat | function | every 1s | main.main (anon func) | in 1s Mar 3 2:15AM | env=local |
// +-----------+----------+----------+-----------------------+--------------------+-----------+
```

## Execution

### <a id="runinbackground"></a>RunInBackground

RunInBackground runs command/exec tasks in a goroutine.

```go
scheduler.New().RunInBackground().Command("noop")
```

## Filters

### <a id="between"></a>Between

Between limits the job to run between the provided HH:MM times (inclusive).

```go
scheduler.New().Between("09:00", "17:00").EveryMinute()
```

### <a id="days"></a>Days

Days limits the job to a specific set of weekdays.

```go
scheduler.New().Days(time.Monday, time.Wednesday, time.Friday).DailyAt("07:00")
```

### <a id="environments"></a>Environments

Environments restricts job registration to specific environment names (e.g. "production", "staging").

```go
scheduler.New().Environments("production").Daily()
```

### <a id="fridays"></a>Fridays

Fridays limits the job to Fridays.

```go
scheduler.New().Fridays().DailyAt("09:00")
```

### <a id="mondays"></a>Mondays

Mondays limits the job to Mondays.

```go
scheduler.New().Mondays().DailyAt("09:00")
```

### <a id="saturdays"></a>Saturdays

Saturdays limits the job to Saturdays.

```go
scheduler.New().Saturdays().DailyAt("09:00")
```

### <a id="skip"></a>Skip

Skip prevents scheduling the job if the provided condition returns true.

```go
enabled := false
scheduler.New().Skip(func() bool { return !enabled }).Daily()
```

### <a id="sundays"></a>Sundays

Sundays limits the job to Sundays.

```go
scheduler.New().Sundays().DailyAt("09:00")
```

### <a id="thursdays"></a>Thursdays

Thursdays limits the job to Thursdays.

```go
scheduler.New().Thursdays().DailyAt("09:00")
```

### <a id="tuesdays"></a>Tuesdays

Tuesdays limits the job to Tuesdays.

```go
scheduler.New().Tuesdays().DailyAt("09:00")
```

### <a id="unlessbetween"></a>UnlessBetween

UnlessBetween prevents the job from running between the provided HH:MM times.

```go
scheduler.New().UnlessBetween("22:00", "06:00").EveryMinute()
```

### <a id="wednesdays"></a>Wednesdays

Wednesdays limits the job to Wednesdays.

```go
scheduler.New().Wednesdays().DailyAt("09:00")
```

### <a id="weekdays"></a>Weekdays

Weekdays limits the job to run only on weekdays (Mon-Fri).

```go
scheduler.New().Weekdays().DailyAt("09:00")
```

### <a id="weekends"></a>Weekends

Weekends limits the job to run only on weekends (Sat-Sun).

```go
scheduler.New().Weekends().DailyAt("10:00")
```

### <a id="when"></a>When

When only schedules the job if the provided condition returns true.

```go
flag := true
scheduler.New().When(func() bool { return flag }).Daily()
```

## Hooks

### <a id="after"></a>After

After sets a hook to run after task execution.

```go
scheduler.New().After(func(context.Context) {}).Daily()
```

### <a id="before"></a>Before

Before sets a hook to run before task execution.

```go
scheduler.New().Before(func(context.Context) {}).Daily()
```

### <a id="onfailure"></a>OnFailure

OnFailure sets a hook to run after failed task execution.

```go
scheduler.New().OnFailure(func(context.Context, error) {}).Daily()
```

### <a id="onsuccess"></a>OnSuccess

OnSuccess sets a hook to run after successful task execution.

```go
scheduler.New().OnSuccess(func(context.Context) {}).Daily()
```

## Interop

### <a id="gocronscheduler"></a>GocronScheduler

GocronScheduler returns the underlying gocron scheduler for advanced integration.
Prefer the fluent scheduler API for typical use-cases.

## Intervals

### <a id="every"></a>Every

Every schedules a job to run every X seconds, minutes, or hours.

```go
scheduler.New().Every(10).Minutes()
```

### <a id="everyduration"></a>EveryDuration

EveryDuration schedules a duration-based interval job builder.

### <a id="everyfifteenminutes"></a>EveryFifteenMinutes

EveryFifteenMinutes schedules the job to run every 15 minutes.

```go
scheduler.New().EveryFifteenMinutes().Do(func(context.Context) error { return nil })
```

### <a id="everyfifteenseconds"></a>EveryFifteenSeconds

EveryFifteenSeconds schedules the job to run every 15 seconds.

```go
scheduler.New().EveryFifteenSeconds().Do(func(context.Context) error { return nil })
```

### <a id="everyfiveminutes"></a>EveryFiveMinutes

EveryFiveMinutes schedules the job to run every 5 minutes.

```go
scheduler.New().EveryFiveMinutes().Do(func(context.Context) error { return nil })
```

### <a id="everyfiveseconds"></a>EveryFiveSeconds

EveryFiveSeconds schedules the job to run every 5 seconds.

```go
scheduler.New().EveryFiveSeconds().Do(func(context.Context) error { return nil })
```

### <a id="everyfourhours"></a>EveryFourHours

EveryFourHours schedules the job to run every four hours at the specified minute.

```go
scheduler.New().EveryFourHours(25)
```

### <a id="everyfourminutes"></a>EveryFourMinutes

EveryFourMinutes schedules the job to run every 4 minutes.

```go
scheduler.New().EveryFourMinutes().Do(func(context.Context) error { return nil })
```

### <a id="everyminute"></a>EveryMinute

EveryMinute schedules the job to run every 1 minute.

```go
scheduler.New().EveryMinute().Do(func(context.Context) error { return nil })
```

### <a id="everyoddhour"></a>EveryOddHour

EveryOddHour schedules the job to run every odd-numbered hour at the specified minute.

```go
scheduler.New().EveryOddHour(10)
```

### <a id="everysecond"></a>EverySecond

EverySecond schedules the job to run every 1 second.

```go
scheduler.New().EverySecond().Do(func(context.Context) error { return nil })
```

### <a id="everysixhours"></a>EverySixHours

EverySixHours schedules the job to run every six hours at the specified minute.

```go
scheduler.New().EverySixHours(30)
```

### <a id="everytenminutes"></a>EveryTenMinutes

EveryTenMinutes schedules the job to run every 10 minutes.

```go
scheduler.New().EveryTenMinutes().Do(func(context.Context) error { return nil })
```

### <a id="everytenseconds"></a>EveryTenSeconds

EveryTenSeconds schedules the job to run every 10 seconds.

```go
scheduler.New().EveryTenSeconds().Do(func(context.Context) error { return nil })
```

### <a id="everythirtyminutes"></a>EveryThirtyMinutes

EveryThirtyMinutes schedules the job to run every 30 minutes.

```go
scheduler.New().EveryThirtyMinutes().Do(func(context.Context) error { return nil })
```

### <a id="everythirtyseconds"></a>EveryThirtySeconds

EveryThirtySeconds schedules the job to run every 30 seconds.

```go
scheduler.New().EveryThirtySeconds().Do(func(context.Context) error { return nil })
```

### <a id="everythreehours"></a>EveryThreeHours

EveryThreeHours schedules the job to run every three hours at the specified minute.

```go
scheduler.New().EveryThreeHours(20)
```

### <a id="everythreeminutes"></a>EveryThreeMinutes

EveryThreeMinutes schedules the job to run every 3 minutes.

```go
scheduler.New().EveryThreeMinutes().Do(func(context.Context) error { return nil })
```

### <a id="everytwentyseconds"></a>EveryTwentySeconds

EveryTwentySeconds schedules the job to run every 20 seconds.

```go
scheduler.New().EveryTwentySeconds().Do(func(context.Context) error { return nil })
```

### <a id="everytwohours"></a>EveryTwoHours

EveryTwoHours schedules the job to run every two hours at the specified minute.

```go
scheduler.New().EveryTwoHours(15)
```

### <a id="everytwominutes"></a>EveryTwoMinutes

EveryTwoMinutes schedules the job to run every 2 minutes.

```go
scheduler.New().EveryTwoMinutes().Do(func(context.Context) error { return nil })
```

### <a id="everytwoseconds"></a>EveryTwoSeconds

EveryTwoSeconds schedules the job to run every 2 seconds.

```go
scheduler.New().EveryTwoSeconds().Do(func(context.Context) error { return nil })
```

### <a id="hourly"></a>Hourly

Hourly schedules the job to run every hour.

```go
scheduler.New().Hourly().Do(func(context.Context) error { return nil })
```

### <a id="hourlyat"></a>HourlyAt

HourlyAt schedules the job to run every hour at the specified minute.

```go
scheduler.New().HourlyAt(5)
```

### <a id="hours"></a>Hours

Hours schedules the job to run every X hours.

```go
scheduler.New().Every(6).Hours()
```

### <a id="minutes"></a>Minutes

Minutes schedules the job to run every X minutes.

```go
scheduler.New().Every(15).Minutes()
```

### <a id="seconds"></a>Seconds

Seconds schedules the job to run every X seconds.

```go
scheduler.New().Every(3).Seconds().Do(func(context.Context) error { return nil })
```

## Lifecycle

### <a id="shutdown"></a>Shutdown

Shutdown gracefully shuts down the underlying scheduler.

```go
s := scheduler.New()
_ = s.Shutdown()
```

### <a id="start"></a>Start

Start starts the underlying scheduler.

```go
s := scheduler.New()
s.Start()
```

### <a id="stop"></a>Stop

Stop gracefully shuts down the scheduler.

```go
s := scheduler.New()
_ = s.Stop()
```

## Locking

### <a id="newcachelocker"></a>NewCacheLocker

NewCacheLocker creates a CacheLocker with a cache lock client and TTL.
The ttl is a lease duration: when it expires, another worker may acquire the
same lock key. For long-running jobs, choose ttl >= worst-case runtime plus a
safety buffer. If your runtime can exceed ttl, prefer a renewing/heartbeat lock strategy.

_Example: use an in-memory cache driver_

```go
client := cache.NewCache(cache.NewMemoryStore(context.Background()))
locker := scheduler.NewCacheLocker(client, 10*time.Minute)
_, _ = locker.Lock(context.Background(), "job")
```

_Example: use the Redis cache driver_

```go
redisStore := rediscache.New(rediscache.Config{
	Addr: "127.0.0.1:6379",
})
redisClient := cache.NewCache(redisStore)
redisLocker := scheduler.NewCacheLocker(redisClient, 10*time.Minute)
_, _ = redisLocker.Lock(context.Background(), "job")
```

### <a id="newredislocker"></a>NewRedisLocker

NewRedisLocker creates a RedisLocker with a client and TTL.
The ttl is a lease duration: when it expires, another worker may acquire the
same lock key. For long-running jobs, choose ttl >= worst-case runtime plus a
safety buffer. If your runtime can exceed ttl, prefer a renewing/heartbeat lock strategy.

```go
client := redis.NewClient(&redis.Options{}) // replace with your client
locker := scheduler.NewRedisLocker(client, 10*time.Minute)
_, _ = locker.Lock(context.Background(), "job")
```

## Metadata

### <a id="jobmetadata"></a>JobMetadata

JobMetadata returns a copy of the tracked job metadata keyed by job ID.

```go
b := scheduler.New().EverySecond().Do(func(context.Context) error { return nil })
for id, meta := range b.JobMetadata() {
	_ = id
	_ = meta.Name
}
```

### <a id="jobsinfo"></a>JobsInfo

JobsInfo returns a stable, sorted snapshot of all known job metadata.
This is a facade-friendly list form of JobMetadata including paused state.

```go
s := scheduler.New()
s.EverySecond().Name("heartbeat").Do(func(context.Context) error { return nil })
for _, job := range s.JobsInfo() {
	_ = job.ID
	_ = job.Name
	_ = job.Paused
}
```

### <a id="name"></a>Name

Name sets an explicit job name.

```go
scheduler.New().Name("cache:refresh").HourlyAt(15)
```

## Runtime control

### <a id="isjobpaused"></a>IsJobPaused

IsJobPaused reports whether a specific job is paused.

### <a id="ispausedall"></a>IsPausedAll

IsPausedAll reports whether global pause is enabled.

### <a id="observe"></a>Observe

Observe registers a lifecycle observer for all scheduled jobs.
Events are emitted consistently across Do, Command, and Exec jobs.

```go
s := scheduler.New()
s.Observe(scheduler.JobObserverFunc(func(event scheduler.JobEvent) {
	if event.Type == scheduler.JobSkipped && event.Reason == "paused" {
		fmt.Println("skipped: paused")
	}
}))
```

### <a id="pauseall"></a>PauseAll

PauseAll pauses execution for all scheduled jobs without removing them.
This is universal across Do, Command, and Exec jobs.
RunNow calls are skipped while pause is active.

```go
s := scheduler.New()
_ = s.PauseAll()
```

### <a id="pausejob"></a>PauseJob

PauseJob pauses execution for a specific scheduled job.
RunNow calls for that job are skipped while paused.

```go
s := scheduler.New()
b := s.EverySecond().Name("heartbeat").Do(func(context.Context) error { return nil })
_ = s.PauseJob(b.Job().ID())
```

### <a id="resumeall"></a>ResumeAll

ResumeAll resumes execution for all paused jobs.

```go
s := scheduler.New()
_ = s.ResumeAll()
```

### <a id="resumejob"></a>ResumeJob

ResumeJob resumes a paused job by ID.

```go
s := scheduler.New()
b := s.EverySecond().Name("heartbeat").Do(func(context.Context) error { return nil })
_ = s.ResumeJob(b.Job().ID())
```

## State management

### <a id="retainstate"></a>RetainState

RetainState allows the job to retain its state after execution.

```go
builder := scheduler.New().EverySecond().RetainState()
builder.Do(func(context.Context) error { return nil })
builder.Do(func(context.Context) error { return nil })
```

## Triggers

### <a id="cron"></a>Cron

Cron sets the cron expression for the job.

```go
builder := scheduler.New().Cron("15 3 * * *")
fmt.Println(builder.CronExpr())
// Output: 15 3 * * *
```

### <a id="do"></a>Do

Do schedules the job with the provided task function.

```go
scheduler.New().Name("cleanup").Cron("0 0 * * *").Do(func(context.Context) error { return nil })
```
<!-- api:embed:end -->
