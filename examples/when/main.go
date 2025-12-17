//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// When only schedules the job if the provided condition returns true.

	// Example: guard scheduling with a flag
	flag := true
	scheduler.NewJobBuilder(nil).
		When(func() bool { return flag }).
		Daily()
}
