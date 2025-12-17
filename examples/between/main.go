//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// Between limits the job to run between the provided HH:MM times (inclusive).

	// Example: allow execution during business hours
	scheduler.NewJobBuilder(nil).
		Between("09:00", "17:00").
		EveryMinute()
}
