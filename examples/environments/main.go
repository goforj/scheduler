//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// Environments restricts job registration to specific environment names (e.g. "production", "staging").

	// Example: only register in production
	scheduler.NewJobBuilder(nil).Environments("production").Daily()
}
