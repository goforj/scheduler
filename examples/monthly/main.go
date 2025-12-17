//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// Monthly schedules the job to run on the first day of each month at midnight.

	// Example: first-of-month billing
	scheduler.NewJobBuilder(nil).Monthly()
}
