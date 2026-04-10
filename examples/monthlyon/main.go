package main

import "github.com/goforj/scheduler/v2"

func main() {
	// MonthlyOn schedules the job to run on a specific day of the month at a given time.

	// Example: run on the 15th of each month
	scheduler.New().MonthlyOn(15, "09:30")
}
