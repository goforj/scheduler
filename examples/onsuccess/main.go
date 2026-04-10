package main

import (
	"context"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// OnSuccess sets a hook to run after successful task execution.

	// Example: record success
	scheduler.New().OnSuccess(func(context.Context) {}).Daily()
}
