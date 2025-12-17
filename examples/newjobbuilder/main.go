//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// NewJobBuilder creates a new JobBuilder with the provided scheduler.

	// Example: create a builder and schedule a heartbeat
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()
	scheduler.NewJobBuilder(s).EverySecond().Do(func() {})
}
