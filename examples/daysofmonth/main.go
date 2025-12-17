//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// DaysOfMonth schedules the job to run on specific days of the month at a given time.

	// Example: run on the 5th and 20th of each month
	scheduler.NewJobBuilder(nil).DaysOfMonth([]int{5, 20}, "07:15")
}
