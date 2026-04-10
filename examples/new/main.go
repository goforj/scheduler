package main

import (
	"context"
	"github.com/goforj/scheduler"
)

func main() {
	// New creates and starts a scheduler facade.
	// It panics only if gocron scheduler construction fails.

	// Example: create scheduler and run a simple interval job
	s := scheduler.New()
	defer s.Stop()
	s.Every(15).Seconds().Do(func(context.Context) error { return nil })
}
