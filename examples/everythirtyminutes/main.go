//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// EveryThirtyMinutes schedules the job to run every 30 minutes.

	// Example: run every thirty minutes
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	scheduler.NewJobBuilder(s).EveryThirtyMinutes().Do(func() {})
}
