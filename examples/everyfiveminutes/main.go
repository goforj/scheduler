package main

import (
	"context"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// EveryFiveMinutes schedules the job to run every 5 minutes.

	// Example: run every five minutes
	scheduler.New().EveryFiveMinutes().Do(func(context.Context) error { return nil })
}
