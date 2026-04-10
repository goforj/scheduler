package main

import "github.com/goforj/scheduler/v2"

func main() {
	// ResumeAll resumes execution for all paused jobs.

	// Example: resume all jobs
	s := scheduler.New()
	_ = s.ResumeAll()
}
