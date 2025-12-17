//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// Hours schedules the job to run every X hours.

	// Example: build an hourly cadence
	scheduler.NewJobBuilder(nil).Every(6).Hours()
}
