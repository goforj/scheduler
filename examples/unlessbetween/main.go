package main

import "github.com/goforj/scheduler/v2"

func main() {
	// UnlessBetween prevents the job from running between the provided HH:MM times.

	// Example: pause execution overnight
	scheduler.New().UnlessBetween("22:00", "06:00").EveryMinute()
}
