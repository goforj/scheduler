//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// EveryFifteenMinutes schedules the job to run every 15 minutes.

	// Example: run every fifteen minutes
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	scheduler.NewJobBuilder(s).EveryFifteenMinutes().Do(func() {})
}
