//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// Before sets a hook to run before task execution.

	// Example: add a before hook
	scheduler.NewJobBuilder(nil).
		Before(func() { println("before") }).
		Daily()
}
