package main

import "github.com/goforj/scheduler"

func main() {
	// Weekends limits the job to run only on weekends (Sat-Sun).

	// Example: weekend-only execution
	scheduler.New().Weekends().DailyAt("10:00")
}
