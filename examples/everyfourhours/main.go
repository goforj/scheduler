//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// EveryFourHours schedules the job to run every four hours at the specified minute.

	// Example: run every four hours
	scheduler.NewJobBuilder(nil).EveryFourHours(25)
}
