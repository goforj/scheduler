package scheduler

import (
	"context"
	"testing"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
)

func newTestSchedulerFacade(t *testing.T) *Scheduler {
	t.Helper()
	clock := clockwork.NewFakeClock()
	raw := newTestScheduler(clock)
	state := newRuntimeState()
	return &Scheduler{
		JobBuilder: newJobBuilderWithState(raw, state),
		s:          raw,
	}
}

func TestListSchedulesBuildsAdminSnapshot(t *testing.T) {
	s := newTestSchedulerFacade(t)
	defer s.Shutdown()

	s.Cron("0 * * * *").Command("hello:world")
	s.Every(2).Minutes().Name("funcJob").Do(func(context.Context) error { return nil })

	entries := s.Admin().ListSchedules()
	require.Len(t, entries, 2)
	require.Equal(t, "funcJob", entries[0].Name)
	require.Equal(t, "function", entries[0].Type)
	require.NotEmpty(t, entries[0].Handler)
	require.Equal(t, "hello:world", entries[1].Name)
	require.Equal(t, "command", entries[1].Type)
	require.Equal(t, "hello:world", entries[1].Handler)
}

func TestPauseResumeScheduleByStringID(t *testing.T) {
	s := newTestSchedulerFacade(t)
	defer s.Shutdown()

	job := s.EveryMinute().Name("heartbeat").Do(sampleHandler).Job()

	require.NoError(t, s.Admin().PauseSchedule(job.ID().String()))
	require.True(t, s.IsJobPaused(job.ID()))

	require.NoError(t, s.Admin().ResumeSchedule(job.ID().String()))
	require.False(t, s.IsJobPaused(job.ID()))
}

func TestRestartScheduleInvalidID(t *testing.T) {
	s := newTestSchedulerFacade(t)
	defer s.Shutdown()

	require.ErrorIs(t, s.Admin().RestartSchedule("nope"), ErrInvalidScheduleID)
}

func TestAdminPauseAllAndResumeAll(t *testing.T) {
	s := newTestSchedulerFacade(t)
	defer s.Shutdown()

	admin := s.Admin()
	require.False(t, admin.IsPausedAll())
	require.NoError(t, admin.PauseAll())
	require.True(t, admin.IsPausedAll())
	require.NoError(t, admin.ResumeAll())
	require.False(t, admin.IsPausedAll())
}
