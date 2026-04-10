package main

import "github.com/goforj/scheduler/v2"

func main() {
	// Hours schedules the job to run every X hours.

	// Example: build an hourly cadence
	scheduler.New().Every(6).Hours()
}
