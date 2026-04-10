package main

import (
	"fmt"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// Observe registers a lifecycle observer for all scheduled jobs.
	// Events are emitted consistently across Do, Command, and Exec jobs.

	// Example: observe paused-skip events
	s := scheduler.New()
	s.Observe(scheduler.JobObserverFunc(func(event scheduler.JobEvent) {
		if event.Type == scheduler.JobSkipped && event.Reason == "paused" {
			fmt.Println("skipped: paused")
		}
	}))
}
