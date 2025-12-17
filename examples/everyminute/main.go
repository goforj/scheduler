//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// EveryMinute schedules the job to run every 1 minute.

	// Example: run a task each minute
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	scheduler.NewJobBuilder(s).EveryMinute().Do(func() {})
}
