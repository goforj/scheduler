//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// Skip prevents scheduling the job if the provided condition returns true.

	// Example: suppress jobs based on a switch
	enabled := false
	scheduler.NewJobBuilder(nil).
		Skip(func() bool { return !enabled }).
		Daily()
}
