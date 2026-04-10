package main

import (
	"context"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// After sets a hook to run after task execution.

	// Example: add an after hook
	scheduler.New().After(func(context.Context) {}).Daily()
}
