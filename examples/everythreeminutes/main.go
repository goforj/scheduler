package main

import (
	"context"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// EveryThreeMinutes schedules the job to run every 3 minutes.

	// Example: run every three minutes
	scheduler.New().EveryThreeMinutes().Do(func(context.Context) error { return nil })
}
