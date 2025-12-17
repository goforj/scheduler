//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// QuarterlyOn schedules the job to run on a specific day of each quarter at a given time.

	// Example: quarterly on a specific day
	scheduler.NewJobBuilder(nil).QuarterlyOn(3, "12:00")
}
