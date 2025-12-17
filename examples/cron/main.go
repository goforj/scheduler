//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"github.com/goforj/scheduler"
)

func main() {
	// Cron sets the cron expression for the job.

	// Example: configure a cron expression
	builder := scheduler.NewJobBuilder(nil).Cron("15 3 * * *")
	fmt.Println(builder.CronExpr())
}
