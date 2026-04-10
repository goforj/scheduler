package main

import (
	"context"
	"github.com/goforj/scheduler"
)

func main() {
	// JobMetadata returns a copy of the tracked job metadata keyed by job ID.

	// Example: inspect scheduled jobs
	b := scheduler.New().EverySecond().Do(func(context.Context) error { return nil })
	for id, meta := range b.JobMetadata() {
		_ = id
		_ = meta.Name
	}
}
