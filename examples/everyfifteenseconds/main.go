package main

import (
	"context"
	"github.com/goforj/scheduler"
)

func main() {
	// EveryFifteenSeconds schedules the job to run every 15 seconds.

	// Example: run at fifteen-second cadence
	scheduler.New().EveryFifteenSeconds().Do(func(context.Context) error { return nil })
}
