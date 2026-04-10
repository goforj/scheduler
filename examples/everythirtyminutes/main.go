package main

import (
	"context"
	"github.com/goforj/scheduler"
)

func main() {
	// EveryThirtyMinutes schedules the job to run every 30 minutes.

	// Example: run every thirty minutes
	scheduler.New().EveryThirtyMinutes().Do(func(context.Context) error { return nil })
}
