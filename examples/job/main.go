package main

import (
	"context"
	"fmt"
	"github.com/goforj/scheduler/v2"
)

func main() {
	// Job returns the last scheduled gocron.Job instance, if available.

	// Example: capture the last job handle
	b := scheduler.New().EverySecond().Do(func(context.Context) error { return nil })
	fmt.Println(b.Job() != nil)
	// Output: true
}
