//go:build ignore
// +build ignore

package main

import "github.com/goforj/scheduler"

func main() {
	// Name sets an explicit job name.

	// Example: label a job for logging
	scheduler.NewJobBuilder(nil).
		Name("cache:refresh").
		HourlyAt(15)
}
