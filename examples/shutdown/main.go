package main

import "github.com/goforj/scheduler"

func main() {
	// Shutdown gracefully shuts down the underlying scheduler.

	// Example: shutdown via underlying method name
	s := scheduler.New()
	_ = s.Shutdown()
}
