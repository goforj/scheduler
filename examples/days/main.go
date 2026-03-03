package main

import (
	"github.com/goforj/scheduler"
	"time"
)

func main() {
	// Days limits the job to a specific set of weekdays.

	// Example: pick custom weekdays
	scheduler.New().Days(time.Monday, time.Wednesday, time.Friday).DailyAt("07:00")
}
