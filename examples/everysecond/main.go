package main

import (
	"context"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// EverySecond schedules the job to run every 1 second.

	// Example: heartbeat job each second
	scheduler.New().EverySecond().Do(func(context.Context) error { return nil })
}
