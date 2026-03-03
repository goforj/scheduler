package main

import "github.com/goforj/scheduler"

func main() {
	// Thursdays limits the job to Thursdays.

	// Example: run only on Thursdays
	scheduler.NewJobBuilder(nil).Thursdays().DailyAt("09:00")
}
