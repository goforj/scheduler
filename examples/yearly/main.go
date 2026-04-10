package main

import "github.com/goforj/scheduler/v2"

func main() {
	// Yearly schedules the job to run on January 1st every year at midnight.

	// Example: yearly trigger
	scheduler.New().Yearly()
}
