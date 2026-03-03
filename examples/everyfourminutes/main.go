package main

import "github.com/goforj/scheduler"

func main() {
	// EveryFourMinutes schedules the job to run every 4 minutes.

	// Example: run every four minutes
	scheduler.New().EveryFourMinutes().Do(func() {})
}
