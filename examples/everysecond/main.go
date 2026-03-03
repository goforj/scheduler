package main

import "github.com/goforj/scheduler"

func main() {
	// EverySecond schedules the job to run every 1 second.

	// Example: heartbeat job each second
	scheduler.New().EverySecond().Do(func() {})
}
