package scheduler

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
)

func TestSchedulerEveryDo(t *testing.T) {
	clock := clockwork.NewFakeClock()

	s, err := NewWithError(gocron.WithClock(clock))
	require.NoError(t, err)
	defer func() { _ = s.Stop() }()

	var called atomic.Bool
	s.Every(1).Seconds().Do(func() { called.Store(true) })

	clock.Advance(time.Second)
	require.Eventually(t, called.Load, 150*time.Millisecond, 10*time.Millisecond)
}

func TestSchedulerCronDo(t *testing.T) {
	s, err := NewWithError()
	require.NoError(t, err)
	require.NoError(t, s.Stop())
}

func TestPauseResumeAllUniversal(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s, err := NewWithError(gocron.WithClock(clock))
	require.NoError(t, err)
	defer func() { _ = s.Stop() }()

	var doCount atomic.Int32
	var runnerCount atomic.Int32

	var mu sync.Mutex
	var skipped []JobEvent
	s.Observe(JobObserverFunc(func(event JobEvent) {
		if event.Type == JobSkipped {
			mu.Lock()
			skipped = append(skipped, event)
			mu.Unlock()
		}
	}))

	runner := CommandRunnerFunc(func(_ context.Context, _ string, _ []string) error {
		runnerCount.Add(1)
		return nil
	})

	s.WithCommandRunner(runner).EverySecond().Do(func() { doCount.Add(1) })
	s.WithCommandRunner(runner).EverySecond().Command("noop")
	s.WithCommandRunner(runner).EverySecond().Exec("/usr/bin/env", "echo", "ok")

	require.NoError(t, s.PauseAll())
	require.True(t, s.IsPausedAll())

	clock.Advance(2 * time.Second)
	time.Sleep(20 * time.Millisecond)

	require.Equal(t, int32(0), doCount.Load())
	require.Equal(t, int32(0), runnerCount.Load())
	require.Len(t, s.Jobs(), 3)

	mu.Lock()
	require.NotEmpty(t, skipped)
	for _, event := range skipped {
		require.Equal(t, string(JobSkipPaused), event.Reason)
	}
	mu.Unlock()

	require.NoError(t, s.ResumeAll())
	require.False(t, s.IsPausedAll())

	clock.Advance(time.Second)
	require.Eventually(t, func() bool {
		return doCount.Load() > 0 && runnerCount.Load() > 0
	}, 250*time.Millisecond, 10*time.Millisecond)
}

func TestPauseResumeSingleJobByTargetKind(t *testing.T) {
	tests := []struct {
		name     string
		schedule func(s *Scheduler, hits *atomic.Int32)
	}{
		{
			name: "function",
			schedule: func(s *Scheduler, hits *atomic.Int32) {
				s.EverySecond().Name("fn").Do(func() { hits.Add(1) })
			},
		},
		{
			name: "command",
			schedule: func(s *Scheduler, hits *atomic.Int32) {
				runner := CommandRunnerFunc(func(_ context.Context, _ string, _ []string) error {
					hits.Add(1)
					return nil
				})
				s.WithCommandRunner(runner).EverySecond().Name("cmd").Command("noop")
			},
		},
		{
			name: "exec",
			schedule: func(s *Scheduler, hits *atomic.Int32) {
				runner := CommandRunnerFunc(func(_ context.Context, _ string, _ []string) error {
					hits.Add(1)
					return nil
				})
				s.WithCommandRunner(runner).EverySecond().Name("exe").Exec("/usr/bin/env", "echo", "ok")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clock := clockwork.NewFakeClock()
			s, err := NewWithError(gocron.WithClock(clock))
			require.NoError(t, err)
			defer func() { _ = s.Stop() }()

			var hits atomic.Int32
			tc.schedule(s, &hits)

			meta := s.JobMetadata()
			require.Len(t, meta, 1)
			var id uuid.UUID
			for jobID := range meta {
				id = jobID
			}

			require.NoError(t, s.PauseJob(id))
			require.True(t, s.IsJobPaused(id))
			clock.Advance(2 * time.Second)
			time.Sleep(20 * time.Millisecond)
			require.Equal(t, int32(0), hits.Load())
			require.Len(t, s.Jobs(), 1)

			require.NoError(t, s.ResumeJob(id))
			require.False(t, s.IsJobPaused(id))
			clock.Advance(time.Second)
			require.Eventually(t, func() bool { return hits.Load() > 0 }, 250*time.Millisecond, 10*time.Millisecond)
		})
	}
}

func TestRunNowWhenPausedIsSkippedUntilResumed(t *testing.T) {
	s, err := NewWithError()
	require.NoError(t, err)
	defer func() { _ = s.Stop() }()

	var hits atomic.Int32
	b := s.EverySecond().Do(func() { hits.Add(1) })
	require.NotNil(t, b.Job())
	id := b.Job().ID()

	require.NoError(t, s.PauseJob(id))
	require.NoError(t, b.Job().RunNow())
	require.Equal(t, int32(0), hits.Load())

	require.NoError(t, s.ResumeJob(id))
	require.False(t, s.IsJobPaused(id))
}

func TestRunNowWhenGloballyPausedIsSkippedUntilResumed(t *testing.T) {
	s, err := NewWithError()
	require.NoError(t, err)
	defer func() { _ = s.Stop() }()

	var hits atomic.Int32
	b := s.EverySecond().Do(func() { hits.Add(1) })
	require.NotNil(t, b.Job())

	require.NoError(t, s.PauseAll())
	require.NoError(t, b.Job().RunNow())
	require.Equal(t, int32(0), hits.Load())

	require.NoError(t, s.ResumeAll())
	require.False(t, s.IsPausedAll())
}

func TestJobsInfoIncludesPausedAndIsStable(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s, err := NewWithError(gocron.WithClock(clock))
	require.NoError(t, err)
	defer func() { _ = s.Stop() }()

	s.EverySecond().Name("b").Do(func() {})
	s.EverySecond().Name("a").Do(func() {})

	jobs := s.JobsInfo()
	require.Len(t, jobs, 2)
	require.Equal(t, "a", jobs[0].Name)
	require.Equal(t, "b", jobs[1].Name)
	for _, job := range jobs {
		require.NotEqual(t, uuid.Nil, job.ID)
		require.NotEmpty(t, job.TargetKind)
		require.False(t, job.Paused)
	}

	meta := s.JobMetadata()
	var pausedID uuid.UUID
	for id, m := range meta {
		if m.Name == "a" {
			pausedID = id
			break
		}
	}
	require.NotEqual(t, uuid.Nil, pausedID)
	require.NoError(t, s.PauseJob(pausedID))

	jobs = s.JobsInfo()
	require.Len(t, jobs, 2)
	require.Equal(t, "a", jobs[0].Name)
	require.True(t, jobs[0].Paused)
	require.False(t, jobs[1].Paused)
}
