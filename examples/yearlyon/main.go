package main

import "github.com/goforj/scheduler/v2"

func main() {
	// YearlyOn schedules the job to run every year on a specific month, day, and time.

	// Example: yearly on a specific date
	scheduler.New().YearlyOn(12, 25, "06:45")
}
