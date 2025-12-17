//go:build ignore
// +build ignore

package main

import (
	"context"
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// WithoutOverlappingWithLocker ensures the job does not run concurrently across distributed systems using the provided locker.

	// Example: use a distributed locker
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
}
