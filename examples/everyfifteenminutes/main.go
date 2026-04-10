package main

import (
	"context"
	"github.com/goforj/scheduler"
)

func main() {
	// EveryFifteenMinutes schedules the job to run every 15 minutes.

	// Example: run every fifteen minutes
	scheduler.New().EveryFifteenMinutes().Do(func(context.Context) error { return nil })
}
