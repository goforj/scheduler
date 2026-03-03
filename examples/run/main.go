package main

import (
	"context"
	"github.com/goforj/scheduler"
)

func main() {
	// Run executes the underlying function.

	// Example: execute the wrapped function
	runner := scheduler.CommandRunnerFunc(func(ctx context.Context, exe string, args []string) error {
		return nil
	})
	_ = runner.Run(context.Background(), "echo", []string{"hi"})
}
