package main

import (
	"context"
	"github.com/goforj/scheduler"
)

func main() {
	// EveryMinute schedules the job to run every 1 minute.

	// Example: run a task each minute
	scheduler.New().EveryMinute().Do(func(context.Context) error { return nil })
}
