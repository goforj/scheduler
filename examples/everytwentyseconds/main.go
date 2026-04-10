package main

import (
	"context"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// EveryTwentySeconds schedules the job to run every 20 seconds.

	// Example: run once every twenty seconds
	scheduler.New().EveryTwentySeconds().Do(func(context.Context) error { return nil })
}
