//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// WeeklyOn schedules the job to run weekly on a specific day of the week and time.
	// Day uses 0 = Sunday through 6 = Saturday.

	// Example: run each Monday at 08:00
	scheduler.NewJobBuilder(nil).WeeklyOn(1, "8:00")
}
