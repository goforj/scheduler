package main

import (
	"fmt"
	"github.com/goforj/scheduler"
)

func main() {
	// Cron schedules a cron-based job builder.

	// Example: configure a cron expression
	builder := scheduler.New().Cron("15 3 * * *")
	fmt.Println(builder.CronExpr())
	// Output: 15 3 * * *
}
