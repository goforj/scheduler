package main

import "github.com/goforj/scheduler/v2"

func main() {
	// EveryTwoHours schedules the job to run every two hours at the specified minute.

	// Example: run every two hours
	scheduler.New().EveryTwoHours(15)
}
