package main

import (
	"context"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// OnFailure sets a hook to run after failed task execution.

	// Example: record failures
	scheduler.New().OnFailure(func(context.Context, error) {}).Daily()
}
