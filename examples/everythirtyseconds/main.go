package main

import "github.com/goforj/scheduler"

func main() {
	// EveryThirtySeconds schedules the job to run every 30 seconds.

	// Example: execute every thirty seconds
	scheduler.New().EveryThirtySeconds().Do(func() {})
}
