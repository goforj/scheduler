package main

import (
	"context"
	"github.com/goforj/scheduler"
	"time"
)

func main() {
	// WithoutOverlapping ensures the job does not run concurrently.

	// Example: prevent overlapping runs of a slow task
	scheduler.New().
		WithoutOverlapping().
		EveryFiveSeconds().
		Do(func(context.Context) error { time.Sleep(7 * time.Second); return nil })
}
