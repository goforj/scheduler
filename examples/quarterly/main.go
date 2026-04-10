package main

import "github.com/goforj/scheduler/v2"

func main() {
	// Quarterly schedules the job to run on the first day of each quarter at midnight.

	// Example: quarterly trigger
	scheduler.New().Quarterly()
}
