package main

import "github.com/goforj/scheduler"

func main() {
	// NewWithError creates and starts a scheduler facade and returns setup errors.

	// Example: construct with explicit error handling
	s, err := scheduler.NewWithError()
	if err != nil {
		panic(err)
	}
	defer s.Stop()
}
