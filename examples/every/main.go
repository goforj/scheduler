//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// Every schedules a job to run every X seconds, minutes, or hours.

	// Example: fluently choose an interval
	scheduler.NewJobBuilder(nil).
		Every(10).
		Minutes()
}
