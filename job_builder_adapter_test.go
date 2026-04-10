package scheduler

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
)

func TestAdapterStartShutdownJobs(t *testing.T) {
	s, err := gocron.NewScheduler()
	require.NoError(t, err)

	adapter := gocronSchedulerAdapter{s: s}
	adapter.Start()
	require.NoError(t, adapter.Shutdown())

	// Jobs should be empty when none scheduled
	require.Empty(t, adapter.Jobs())
}

type stubRunner struct {
	called int
	err    error
	done   chan struct{}
}

func (sr *stubRunner) Run(ctx context.Context, exe string, args []string) error {
	sr.called++
	if sr.done != nil {
		close(sr.done)
	}
	return sr.err
}

func TestWithCommandRunnerAndFailureHook(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	done := make(chan struct{})
	runner := &stubRunner{err: fmt.Errorf("boom"), done: done}
	var failureCalled bool
	var failureErr error

	jb := newJobBuilder(s).
		WithCommandRunner(runner).
		OnFailure(func(_ context.Context, err error) {
			failureCalled = true
			failureErr = err
		}).
		Cron("0 0 * * *").
		Command("hello:world")

	require.NotNil(t, jb.Job())
	require.NoError(t, jb.Job().RunNow())

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("command runner did not run")
	}

	require.Equal(t, 1, runner.called)
	require.True(t, failureCalled)
	require.ErrorIs(t, failureErr, runner.err)
}
