//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// HourlyAt schedules the job to run every hour at the specified minute.

	// Example: run at the 5th minute of each hour
	scheduler.NewJobBuilder(nil).HourlyAt(5)
}
