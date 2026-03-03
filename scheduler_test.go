package scheduler

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
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
