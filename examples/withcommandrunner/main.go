//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"github.com/goforj/scheduler"
)

func main() {
	// WithCommandRunner overrides command execution (default: exec.CommandContext).

	// Example: swap in a custom runner
	runner := scheduler.CommandRunnerFunc(func(_ context.Context, exe string, args []string) error {
		fmt.Println(exe, args)
		return nil
	})

	builder := scheduler.NewJobBuilder(nil).
		WithCommandRunner(runner)
	fmt.Printf("%T\n", builder)
}
