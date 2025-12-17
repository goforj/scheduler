//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// Weekly schedules the job to run once per week on Sunday at midnight.

	// Example: weekly maintenance
	scheduler.NewJobBuilder(nil).Weekly()
}
