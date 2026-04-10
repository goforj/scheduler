package main

import "github.com/goforj/scheduler/v2"

func main() {
	// Stop gracefully shuts down the scheduler.

	// Example: stop the scheduler
	s := scheduler.New()
	_ = s.Stop()
}
