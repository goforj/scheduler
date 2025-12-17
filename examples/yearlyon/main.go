//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// YearlyOn schedules the job to run every year on a specific month, day, and time.

	// Example: yearly on a specific date
	scheduler.NewJobBuilder(nil).YearlyOn(12, 25, "06:45")
}
