package main

import "github.com/goforj/scheduler/v2"

func main() {
	// Name sets an explicit job name.

	// Example: label a job for logging
	scheduler.New().Name("cache:refresh").HourlyAt(15)
}
