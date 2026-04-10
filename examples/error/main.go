package main

import (
	"fmt"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// Error returns the error if any occurred during job scheduling.

	// Example: validate a malformed schedule
	builder := scheduler.New().DailyAt("bad")
	fmt.Println(builder.Error())
	// Output: invalid DailyAt time format: invalid time format (expected HH:MM): "bad"
}
