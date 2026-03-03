package main

import (
	"context"
	"github.com/goforj/scheduler"
)

func main() {
	// WithCommandRunner overrides command execution (default: exec.CommandContext).

	// Example: swap in a custom runner
	runner := scheduler.CommandRunnerFunc(func(_ context.Context, exe string, args []string) error {
		_ = exe
		_ = args
		return nil
	})

	builder := scheduler.New().WithCommandRunner(runner)
	_ = builder
}
