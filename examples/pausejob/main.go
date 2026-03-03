package main

import "github.com/goforj/scheduler"

func main() {
	// PauseJob pauses execution for a specific scheduled job.
	// RunNow calls for that job are skipped while paused.

	// Example: pause one job by ID
	s := scheduler.New()
	b := s.EverySecond().Name("heartbeat").Do(func() {})
	_ = s.PauseJob(b.Job().ID())
}
