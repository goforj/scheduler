//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// EveryThreeMinutes schedules the job to run every 3 minutes.

	// Example: run every three minutes
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	scheduler.NewJobBuilder(s).EveryThreeMinutes().Do(func() {})
}
