//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// JobMetadata returns a copy of the tracked job metadata keyed by job ID.

	// Example: inspect scheduled jobs
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	b := scheduler.NewJobBuilder(s).EverySecond().Do(func() {})
	for id, meta := range b.JobMetadata() {
		_ = id
		_ = meta.Name
	}
}
