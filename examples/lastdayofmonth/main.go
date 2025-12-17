//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// LastDayOfMonth schedules the job to run on the last day of each month at a specific time.

	// Example: run on the last day of the month
	scheduler.NewJobBuilder(nil).LastDayOfMonth("23:30")
}
