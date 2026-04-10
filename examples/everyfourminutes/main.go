package main

import (
	"context"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// EveryFourMinutes schedules the job to run every 4 minutes.

	// Example: run every four minutes
	scheduler.New().EveryFourMinutes().Do(func(context.Context) error { return nil })
}
