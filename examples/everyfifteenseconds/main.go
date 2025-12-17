//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// EveryFifteenSeconds schedules the job to run every 15 seconds.

	// Example: run at fifteen-second cadence
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	scheduler.NewJobBuilder(s).EveryFifteenSeconds().Do(func() {})
}
