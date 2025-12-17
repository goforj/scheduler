//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"github.com/goforj/scheduler"
)

func main() {
	// CronExpr returns the cron expression string configured for this job.

	// Example: inspect the stored cron expression
	builder := scheduler.NewJobBuilder(nil).Cron("0 9 * * *")
	fmt.Println(builder.CronExpr())
}
