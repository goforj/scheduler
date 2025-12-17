//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// Weekends limits the job to run only on weekends (Sat-Sun).

	// Example: weekend-only execution
	scheduler.NewJobBuilder(nil).Weekends().DailyAt("10:00")
}
