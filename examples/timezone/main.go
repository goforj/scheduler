//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// Timezone sets a timezone string for the job (not currently applied to gocron Scheduler).

	// Example: tag jobs with a timezone
	scheduler.NewJobBuilder(nil).
		Timezone("America/New_York").
		Daily()
}
