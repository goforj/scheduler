package main

import "github.com/goforj/scheduler/v2"

func main() {
	// Weekly schedules the job to run once per week on Sunday at midnight.

	// Example: weekly maintenance
	scheduler.New().Weekly()
}
