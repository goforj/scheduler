//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// RunInBackground runs command/exec tasks in a goroutine.

	// Example: allow command jobs to run async
	scheduler.NewJobBuilder(nil).
		RunInBackground().
		Command("noop")
}
