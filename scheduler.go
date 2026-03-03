package scheduler

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
)

// Scheduler is a high-level facade that owns a gocron scheduler instance.
// It starts automatically and exposes low-ceremony entrypoints for jobs.
type Scheduler struct {
	*JobBuilder
	s gocron.Scheduler
}

// New creates and starts a scheduler facade.
// It panics only if gocron scheduler construction fails.
// @group Construction
//
// Example: create scheduler and run a simple interval job
//
//	s := scheduler.New()
//	defer s.Stop()
//	s.Every(15).Seconds().Do(func() {})
func New(options ...gocron.SchedulerOption) *Scheduler {
	s, err := NewWithError(options...)
	if err != nil {
		panic(fmt.Sprintf("scheduler.New failed: %v", err))
	}
	return s
}

// NewWithError creates and starts a scheduler facade and returns setup errors.
// @group Construction
//
// Example: construct with explicit error handling
//
//	s, err := scheduler.NewWithError()
//	if err != nil {
//		panic(err)
//	}
//	defer s.Stop()
func NewWithError(options ...gocron.SchedulerOption) (*Scheduler, error) {
	s, err := gocron.NewScheduler(options...)
	if err != nil {
		return nil, err
	}
	s.Start()
	state := newRuntimeState()
	return &Scheduler{
		JobBuilder: newJobBuilderWithState(s, state),
		s:          s,
	}, nil
}

// Stop gracefully shuts down the scheduler.
// @group Lifecycle
//
// Example: stop the scheduler
//
//	s := scheduler.New()
//	_ = s.Stop()
func (s *Scheduler) Stop() error {
	return s.s.Shutdown()
}

// Start starts the underlying scheduler.
// @group Lifecycle
//
// Example: manually start (auto-started by New/NewWithError)
//
//	s := scheduler.New()
//	s.Start()
func (s *Scheduler) Start() {
	s.s.Start()
}

// Shutdown gracefully shuts down the underlying scheduler.
// @group Lifecycle
//
// Example: shutdown via underlying method name
//
//	s := scheduler.New()
//	_ = s.Shutdown()
func (s *Scheduler) Shutdown() error {
	return s.s.Shutdown()
}

// Jobs returns scheduled jobs from the underlying scheduler.
// @group Diagnostics
func (s *Scheduler) Jobs() []gocron.Job {
	return s.s.Jobs()
}

func (s *Scheduler) Every(interval int) *FluentEvery {
	return newJobBuilderWithState(s.s, s.JobBuilder.state).Every(interval)
}

// EveryDuration schedules a duration-based interval job builder.
// @group Intervals
func (s *Scheduler) EveryDuration(interval time.Duration) *JobBuilder {
	return newJobBuilderWithState(s.s, s.JobBuilder.state).every(interval)
}

func (s *Scheduler) Cron(expr string) *JobBuilder {
	return newJobBuilderWithState(s.s, s.JobBuilder.state).Cron(expr)
}

// GocronScheduler returns the underlying gocron scheduler for advanced integration.
// Prefer the fluent scheduler API for typical use-cases.
// @group Interop
func (s *Scheduler) GocronScheduler() gocron.Scheduler {
	return s.s
}

func (s *Scheduler) Raw() gocron.Scheduler {
	return s.GocronScheduler()
}
