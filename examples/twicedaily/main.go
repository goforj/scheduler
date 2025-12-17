//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// TwiceDaily schedules the job to run daily at two specified hours (e.g., 1 and 13).

	// Example: run two times per day
	scheduler.NewJobBuilder(nil).TwiceDaily(1, 13)
}
