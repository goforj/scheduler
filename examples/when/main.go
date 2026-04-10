package main

import "github.com/goforj/scheduler/v2"

func main() {
	// When only schedules the job if the provided condition returns true.

	// Example: guard scheduling with a flag
	flag := true
	scheduler.New().When(func() bool { return flag }).Daily()
}
