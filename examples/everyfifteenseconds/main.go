package main

import "github.com/goforj/scheduler"

func main() {
	// EveryFifteenSeconds schedules the job to run every 15 seconds.

	// Example: run at fifteen-second cadence
	scheduler.New().EveryFifteenSeconds().Do(func() {})
}
