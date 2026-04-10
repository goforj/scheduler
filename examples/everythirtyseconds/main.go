package main

import (
	"context"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// EveryThirtySeconds schedules the job to run every 30 seconds.

	// Example: execute every thirty seconds
	scheduler.New().EveryThirtySeconds().Do(func(context.Context) error { return nil })
}
