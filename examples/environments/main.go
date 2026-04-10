package main

import "github.com/goforj/scheduler/v2"

func main() {
	// Environments restricts job registration to specific environment names (e.g. "production", "staging").

	// Example: only register in production
	scheduler.New().Environments("production").Daily()
}
