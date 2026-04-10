package main

import "github.com/goforj/scheduler/v2"

func main() {
	// Fridays limits the job to Fridays.

	// Example: run only on Fridays
	scheduler.New().Fridays().DailyAt("09:00")
}
