package main

import "github.com/goforj/scheduler/v2"

func main() {
	// Daily schedules the job to run once per day at midnight.

	// Example: nightly task
	scheduler.New().Daily()
}
