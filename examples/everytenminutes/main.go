package main

import (
	"context"
	"github.com/goforj/scheduler"
)

func main() {
	// EveryTenMinutes schedules the job to run every 10 minutes.

	// Example: run every ten minutes
	scheduler.New().EveryTenMinutes().Do(func(context.Context) error { return nil })
}
