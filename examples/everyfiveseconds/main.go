package main

import "github.com/goforj/scheduler"

func main() {
	// EveryFiveSeconds schedules the job to run every 5 seconds.

	// Example: space out work every five seconds
	scheduler.New().EveryFiveSeconds().Do(func() {})
}
