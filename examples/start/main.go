package main

import "github.com/goforj/scheduler/v2"

func main() {
	// Start starts the underlying scheduler.

	// Example: manually start (auto-started by New/NewWithError)
	s := scheduler.New()
	s.Start()
}
