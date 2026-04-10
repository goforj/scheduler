package main

import (
	"context"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// EveryTwoMinutes schedules the job to run every 2 minutes.

	// Example: job that runs every two minutes
	scheduler.New().EveryTwoMinutes().Do(func(context.Context) error { return nil })
}
