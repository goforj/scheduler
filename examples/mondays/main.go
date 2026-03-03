package main

import "github.com/goforj/scheduler"

func main() {
	// Mondays limits the job to Mondays.

	// Example: run only on Mondays
	scheduler.NewJobBuilder(nil).Mondays().DailyAt("09:00")
}
