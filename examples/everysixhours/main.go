package main

import "github.com/goforj/scheduler/v2"

func main() {
	// EverySixHours schedules the job to run every six hours at the specified minute.

	// Example: run every six hours
	scheduler.New().EverySixHours(30)
}
