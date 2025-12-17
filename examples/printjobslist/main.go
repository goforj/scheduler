//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// PrintJobsList renders and prints the scheduler job table to stdout.

	// Example: print current jobs
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	scheduler.NewJobBuilder(s).
		EverySecond().
		Name("heartbeat").
		Do(func() {})

	scheduler.NewJobBuilder(s).PrintJobsList()
}
