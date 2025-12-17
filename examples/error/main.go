//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"github.com/goforj/scheduler"
)

func main() {
	// Error returns the error if any occurred during job scheduling.

	// Example: validate a malformed schedule
	builder := scheduler.NewJobBuilder(nil).DailyAt("bad")
	fmt.Println(builder.Error())
}
