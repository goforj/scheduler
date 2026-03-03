package main

import (
	"github.com/goforj/scheduler"
	"time"
)

func main() {
	// WithoutOverlapping ensures the job does not run concurrently.

	// Example: prevent overlapping runs of a slow task
	scheduler.New().
		WithoutOverlapping().
		EveryFiveSeconds().
		Do(func() { time.Sleep(7 * time.Second) })
}
