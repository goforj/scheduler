//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// Minutes schedules the job to run every X minutes.

	// Example: chain a minute-based interval
	scheduler.NewJobBuilder(nil).Every(15).Minutes()
}
