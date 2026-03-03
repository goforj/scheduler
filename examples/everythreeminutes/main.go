package main

import "github.com/goforj/scheduler"

func main() {
	// EveryThreeMinutes schedules the job to run every 3 minutes.

	// Example: run every three minutes
	scheduler.New().EveryThreeMinutes().Do(func() {})
}
