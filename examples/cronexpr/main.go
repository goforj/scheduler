package main

import (
	"fmt"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// CronExpr returns the cron expression string configured for this job.

	// Example: inspect the stored cron expression
	builder := scheduler.New().Cron("0 9 * * *")
	fmt.Println(builder.CronExpr())
	// Output: 0 9 * * *
}
