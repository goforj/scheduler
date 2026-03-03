package main

import "github.com/goforj/scheduler"

func main() {
	// Tuesdays limits the job to Tuesdays.

	// Example: run only on Tuesdays
	scheduler.New().Tuesdays().DailyAt("09:00")
}
