package main

import "github.com/goforj/scheduler"

func main() {
	// EveryTwoMinutes schedules the job to run every 2 minutes.

	// Example: job that runs every two minutes
	scheduler.New().EveryTwoMinutes().Do(func() {})
}
