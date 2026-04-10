package main

import "github.com/goforj/scheduler/v2"

func main() {
	// EveryOddHour schedules the job to run every odd-numbered hour at the specified minute.

	// Example: run every odd hour
	scheduler.New().EveryOddHour(10)
}
