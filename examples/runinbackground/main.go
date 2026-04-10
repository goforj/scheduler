package main

import "github.com/goforj/scheduler/v2"

func main() {
	// RunInBackground runs command/exec tasks in a goroutine.

	// Example: allow command jobs to run async
	scheduler.New().RunInBackground().Command("noop")
}
