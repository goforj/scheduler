//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// EveryThreeHours schedules the job to run every three hours at the specified minute.

	// Example: run every three hours
	scheduler.NewJobBuilder(nil).EveryThreeHours(20)
}
