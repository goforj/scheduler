package main

import "github.com/goforj/scheduler/v2"

func main() {
	// Skip prevents scheduling the job if the provided condition returns true.

	// Example: suppress jobs based on a switch
	enabled := false
	scheduler.New().Skip(func() bool { return !enabled }).Daily()
}
