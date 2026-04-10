package main

import "github.com/goforj/scheduler/v2"

func main() {
	// DailyAt schedules the job to run daily at a specific time (e.g., "13:00").

	// Example: run at lunch time daily
	scheduler.New().DailyAt("12:30")
}
