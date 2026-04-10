package main

import (
	"context"
	"github.com/goforj/scheduler"
)

func main() {
	// JobsInfo returns a stable, sorted snapshot of all known job metadata.
	// This is a facade-friendly list form of JobMetadata including paused state.

	// Example: iterate jobs for UI rendering
	s := scheduler.New()
	s.EverySecond().Name("heartbeat").Do(func(context.Context) error { return nil })
	for _, job := range s.JobsInfo() {
		_ = job.ID
		_ = job.Name
		_ = job.Paused
	}
}
