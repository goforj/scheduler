package main

import "github.com/goforj/scheduler"

func main() {
	// Wednesdays limits the job to Wednesdays.

	// Example: run only on Wednesdays
	scheduler.NewJobBuilder(nil).Wednesdays().DailyAt("09:00")
}
