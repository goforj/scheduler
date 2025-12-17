//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// Yearly schedules the job to run on January 1st every year at midnight.

	// Example: yearly trigger
	scheduler.NewJobBuilder(nil).Yearly()
}
