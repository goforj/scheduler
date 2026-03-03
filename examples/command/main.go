package main

import "github.com/goforj/scheduler"

func main() {
	// Command executes the current binary with the given subcommand and variadic args.
	// It does not run arbitrary system executables; use Exec for that.

	// Example: run a CLI subcommand on schedule
	scheduler.New().Cron("0 0 * * *").Command("jobs:purge", "--force")
}
