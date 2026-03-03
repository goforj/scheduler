package main

import "github.com/goforj/scheduler"

func main() {
	// EveryFourHours schedules the job to run every four hours at the specified minute.

	// Example: run every four hours
	scheduler.New().EveryFourHours(25)
}
