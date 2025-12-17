//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
	"time"
)

func main() {
	// WithoutOverlapping ensures the job does not run concurrently.

	// Example: prevent overlapping runs of a slow task
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	scheduler.NewJobBuilder(s).
		WithoutOverlapping().
		EveryFiveSeconds().
		Do(func() { time.Sleep(7 * time.Second) })
}
