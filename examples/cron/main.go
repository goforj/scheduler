package main

import (
	"fmt"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// Cron sets the cron expression for the job.

	// Example: configure a cron expression
	builder := scheduler.New().Cron("15 3 * * *")
	fmt.Println(builder.CronExpr())
	// Output: 15 3 * * *
}
