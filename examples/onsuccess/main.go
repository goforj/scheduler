//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// OnSuccess sets a hook to run after successful task execution.

	// Example: record success
	scheduler.NewJobBuilder(nil).
		OnSuccess(func() { println("success") }).
		Daily()
}
