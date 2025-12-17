//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// OnFailure sets a hook to run after failed task execution.

	// Example: record failures
	scheduler.NewJobBuilder(nil).
		OnFailure(func() { println("failure") }).
		Daily()
}
