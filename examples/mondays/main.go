package main

import "github.com/goforj/scheduler/v2"

func main() {
	// Mondays limits the job to Mondays.

	// Example: run only on Mondays
	scheduler.New().Mondays().DailyAt("09:00")
}
