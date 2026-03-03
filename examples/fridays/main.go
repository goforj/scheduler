package main

import "github.com/goforj/scheduler"

func main() {
	// Fridays limits the job to Fridays.

	// Example: run only on Fridays
	scheduler.NewJobBuilder(nil).Fridays().DailyAt("09:00")
}
