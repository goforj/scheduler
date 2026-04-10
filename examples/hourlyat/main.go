package main

import "github.com/goforj/scheduler/v2"

func main() {
	// HourlyAt schedules the job to run every hour at the specified minute.

	// Example: run at the 5th minute of each hour
	scheduler.New().HourlyAt(5)
}
