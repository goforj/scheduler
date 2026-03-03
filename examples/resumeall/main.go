package main

import "github.com/goforj/scheduler"

func main() {
	// ResumeAll resumes execution for all paused jobs.

	// Example: resume all jobs
	s := scheduler.New()
	_ = s.ResumeAll()
}
