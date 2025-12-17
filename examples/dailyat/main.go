//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// DailyAt schedules the job to run daily at a specific time (e.g., "13:00").

	// Example: run at lunch time daily
	scheduler.NewJobBuilder(nil).DailyAt("12:30")
}
