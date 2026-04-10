package main

import "github.com/goforj/scheduler/v2"

func main() {
	// LastDayOfMonth schedules the job to run on the last day of each month at a specific time.

	// Example: run on the last day of the month
	scheduler.New().LastDayOfMonth("23:30")
}
