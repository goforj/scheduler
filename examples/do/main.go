//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// Do schedules the job with the provided task function.

	// Example: create a named cron job
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	scheduler.NewJobBuilder(s).
	Name("cleanup").
	Cron("0 0 * * *").
	Do(func() {})
}
