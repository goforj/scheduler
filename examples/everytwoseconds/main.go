package main

import (
	"context"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// EveryTwoSeconds schedules the job to run every 2 seconds.

	// Example: throttle a task to two seconds
	scheduler.New().EveryTwoSeconds().Do(func(context.Context) error { return nil })
}
