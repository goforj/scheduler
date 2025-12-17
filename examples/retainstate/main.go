//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// RetainState allows the job to retain its state after execution.

	// Example: reuse interval configuration for multiple jobs
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	builder := scheduler.NewJobBuilder(s).EverySecond().RetainState()
	builder.Do(func() {})
	builder.Do(func() {})
}
