package main

import (
	"context"
	"github.com/goforj/scheduler"
)

func main() {
	// After sets a hook to run after task execution.

	// Example: add an after hook
	scheduler.New().After(func(context.Context) {}).Daily()
}
