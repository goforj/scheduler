package main

import (
	"context"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// Seconds schedules the job to run every X seconds.

	// Example: run a task every few seconds
	scheduler.New().Every(3).Seconds().Do(func(context.Context) error { return nil })
}
