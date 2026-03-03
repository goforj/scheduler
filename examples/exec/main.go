package main

import "github.com/goforj/scheduler"

func main() {
	// Exec runs an external executable with variadic args.

	// Example: run a system executable on schedule
	scheduler.New().Cron("0 0 * * *").Exec("/usr/bin/env", "echo", "hello")
}
