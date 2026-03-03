package main

import "github.com/goforj/scheduler"

func main() {
	// Do schedules the job with the provided task function.

	// Example: create a named cron job
	scheduler.New().Name("cleanup").Cron("0 0 * * *").Do(func() {})
}
