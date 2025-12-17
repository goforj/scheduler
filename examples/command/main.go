//go:build ignore
// +build ignore

package main

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// Command executes the current binary with the given subcommand and variadic args.

	// Example: run a CLI subcommand on schedule
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	scheduler.NewJobBuilder(s).
		Cron("0 0 * * *").
		Command("jobs:purge", "--force")
}
