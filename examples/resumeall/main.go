package main

import "github.com/goforj/scheduler/v2"

func main() {
	// ResumeAll restarts job execution for all schedules.

	// Example: resume all jobs
	s := scheduler.New()
	_ = s.ResumeAll()
}
