//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// TwiceMonthly schedules the job to run on two specific days of the month at the given time.

	// Example: run on two days each month
	scheduler.NewJobBuilder(nil).TwiceMonthly(1, 15, "10:00")
}
