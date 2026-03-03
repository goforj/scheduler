package main

import "github.com/goforj/scheduler"

func main() {
	// RetainState allows the job to retain its state after execution.

	// Example: reuse interval configuration for multiple jobs
	builder := scheduler.New().EverySecond().RetainState()
	builder.Do(func() {})
	builder.Do(func() {})
}
