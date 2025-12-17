//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// EveryTwoSeconds schedules the job to run every 2 seconds.

	// Example: throttle a task to two seconds
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	scheduler.NewJobBuilder(s).EveryTwoSeconds().Do(func() {})
}
