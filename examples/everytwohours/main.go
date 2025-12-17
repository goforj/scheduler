//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// EveryTwoHours schedules the job to run every two hours at the specified minute.

	// Example: run every two hours
	scheduler.NewJobBuilder(nil).EveryTwoHours(15)
}
