//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// EveryFiveSeconds schedules the job to run every 5 seconds.

	// Example: space out work every five seconds
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	scheduler.NewJobBuilder(s).EveryFiveSeconds().Do(func() {})
}
