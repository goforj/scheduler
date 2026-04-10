package main

import (
	"context"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// EveryFiveSeconds schedules the job to run every 5 seconds.

	// Example: space out work every five seconds
	scheduler.New().EveryFiveSeconds().Do(func(context.Context) error { return nil })
}
