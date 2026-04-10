package main

import "github.com/goforj/scheduler/v2"

func main() {
	// PauseAll pauses execution for all scheduled jobs without removing them.
	// This is universal across Do, Command, and Exec jobs.
	// RunNow calls are skipped while pause is active.

	// Example: pause all jobs
	s := scheduler.New()
	_ = s.PauseAll()
}
