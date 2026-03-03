package main

import "github.com/goforj/scheduler"

func main() {
	// PauseAll pauses execution for all scheduled jobs without removing them.
	// This is universal across Do, Command, and Exec jobs.

	// Example: pause all jobs
	s := scheduler.New()
	_ = s.PauseAll()
}
