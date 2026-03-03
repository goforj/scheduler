package main

import "github.com/goforj/scheduler"

func main() {
	// Saturdays limits the job to Saturdays.

	// Example: run only on Saturdays
	scheduler.NewJobBuilder(nil).Saturdays().DailyAt("09:00")
}
