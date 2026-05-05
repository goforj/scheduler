package main

import "github.com/goforj/scheduler/v2"

func main() {
	// PauseAll halts job execution without removing schedule definitions.

	// Example: pause all jobs
	s := scheduler.New()
	_ = s.PauseAll()
}
