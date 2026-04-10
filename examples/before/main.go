package main

import (
	"context"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// Before sets a hook to run before task execution.

	// Example: add a before hook
	scheduler.New().Before(func(context.Context) {}).Daily()
}
