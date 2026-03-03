package main

import "github.com/goforj/scheduler"

func main() {
	// EveryTenSeconds schedules the job to run every 10 seconds.

	// Example: poll every ten seconds
	scheduler.New().EveryTenSeconds().Do(func() {})
}
