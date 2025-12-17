//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// Quarterly schedules the job to run on the first day of each quarter at midnight.

	// Example: quarterly trigger
	scheduler.NewJobBuilder(nil).Quarterly()
}
