//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// After sets a hook to run after task execution.

	// Example: add an after hook
	scheduler.NewJobBuilder(nil).
		After(func() { println("after") }).
		Daily()
}
