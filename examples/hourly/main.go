package main

import (
	"context"
	"github.com/goforj/scheduler"
)

func main() {
	// Hourly schedules the job to run every hour.

	// Example: run something hourly
	scheduler.New().Hourly().Do(func(context.Context) error { return nil })
}
