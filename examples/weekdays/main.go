package main

import "github.com/goforj/scheduler"

func main() {
	// Weekdays limits the job to run only on weekdays (Mon-Fri).

	// Example: weekday-only execution
	scheduler.New().Weekdays().DailyAt("09:00")
}
