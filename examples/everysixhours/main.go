//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// EverySixHours schedules the job to run every six hours at the specified minute.

	// Example: run every six hours
	scheduler.NewJobBuilder(nil).EverySixHours(30)
}
