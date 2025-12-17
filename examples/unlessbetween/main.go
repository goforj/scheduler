//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// UnlessBetween prevents the job from running between the provided HH:MM times.

	// Example: pause execution overnight
	scheduler.NewJobBuilder(nil).
		UnlessBetween("22:00", "06:00").
		EveryMinute()
}
