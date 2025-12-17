//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// EveryThirtySeconds schedules the job to run every 30 seconds.

	// Example: execute every thirty seconds
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	scheduler.NewJobBuilder(s).EveryThirtySeconds().Do(func() {})
}
