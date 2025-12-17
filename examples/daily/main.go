//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// Daily schedules the job to run once per day at midnight.

	// Example: nightly task
	scheduler.NewJobBuilder(nil).Daily()
}
