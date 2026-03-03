package main

import "github.com/goforj/scheduler"

func main() {
	// Sundays limits the job to Sundays.

	// Example: run only on Sundays
	scheduler.New().Sundays().DailyAt("09:00")
}
