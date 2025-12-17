//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// EveryTenSeconds schedules the job to run every 10 seconds.

	// Example: poll every ten seconds
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	scheduler.NewJobBuilder(s).EveryTenSeconds().Do(func() {})
}
