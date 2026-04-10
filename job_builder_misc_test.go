package scheduler

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
)

func TestWithNowFuncAndName(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	fixed := time.Date(2025, time.January, 1, 10, 0, 0, 0, time.UTC)
	jb := newJobBuilder(s).WithNowFunc(func() time.Time { return fixed }).Name("custom")
	jb.EveryMinute().Do(func(context.Context) error { return nil })

	require.Equal(t, fixed, jb.now())
	require.Equal(t, "custom", jb.Job().Name())
}

func TestRunInBackgroundCommand(t *testing.T) {
	clock := clockwork.NewFakeClock()
	s := newTestScheduler(clock)
	defer s.Shutdown()

	done := make(chan struct{})
	runner := &stubRunner{done: done}

	jb := newJobBuilder(s).
		RunInBackground().
		WithCommandRunner(runner).
		Cron("0 * * * *").
		Command("demo:task")

	require.NotNil(t, jb.Job())
	require.NoError(t, jb.Job().RunNow())

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("background command did not complete")
	}
}

func TestExecCommandRunner(t *testing.T) {
	tmpDir := t.TempDir()
	script := filepath.Join(tmpDir, "echo.sh")
	err := os.WriteFile(script, []byte("#!/bin/sh\necho hello"), 0o755)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = (execCommandRunner{}).Run(ctx, script, nil)
	require.NoError(t, err)
}

func TestLocationFallback(t *testing.T) {
	jb := newJobBuilder(nil).Timezone("bad/zone")
	require.Equal(t, time.Local, jb.location())
}

func TestLockError(t *testing.T) {
	err := (&lockError{msg: "fail"}).Error()
	require.Equal(t, "fail", err)
}
