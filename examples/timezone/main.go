package main

import "github.com/goforj/scheduler/v2"

func main() {
	// Timezone sets a timezone string for the job (not currently applied to gocron Scheduler).

	// Example: tag jobs with a timezone
	scheduler.New().Timezone("America/New_York").Daily()
}
