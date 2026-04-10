package main

import "github.com/goforj/scheduler/v2"

func main() {
	// Saturdays limits the job to Saturdays.

	// Example: run only on Saturdays
	scheduler.New().Saturdays().DailyAt("09:00")
}
