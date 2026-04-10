package main

import "github.com/goforj/scheduler/v2"

func main() {
	// WeeklyOn schedules the job to run weekly on a specific day of the week and time.
	// Day uses 0 = Sunday through 6 = Saturday.

	// Example: run each Monday at 08:00
	scheduler.New().WeeklyOn(1, "8:00")
}
