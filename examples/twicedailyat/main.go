//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// TwiceDailyAt schedules the job to run daily at two specified times (e.g., 1:15 and 13:15).

	// Example: run twice daily at explicit minutes
	scheduler.NewJobBuilder(nil).TwiceDailyAt(1, 13, 15)
}
