package main

import (
	"context"
	"github.com/goforj/scheduler"
)

func main() {
	// ResumeJob resumes a paused job by ID.

	// Example: resume one job by ID
	s := scheduler.New()
	b := s.EverySecond().Name("heartbeat").Do(func(context.Context) error { return nil })
	_ = s.ResumeJob(b.Job().ID())
}
