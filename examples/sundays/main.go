package main

import "github.com/goforj/scheduler"

func main() {
	// Sundays limits the job to Sundays.

	// Example: run only on Sundays
	scheduler.NewJobBuilder(nil).Sundays().DailyAt("09:00")
}
