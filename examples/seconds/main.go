//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// Seconds schedules the job to run every X seconds.

	// Example: run a task every few seconds
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	scheduler.NewJobBuilder(s).
		Every(3).
		Seconds().
		Do(func() {})
}
