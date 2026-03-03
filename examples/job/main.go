package main

import (
	"fmt"
	"github.com/goforj/scheduler"
)

func main() {
	// Job returns the last scheduled gocron.Job instance, if available.

	// Example: capture the last job handle
	b := scheduler.New().EverySecond().Do(func() {})
	fmt.Println(b.Job() != nil)
	// Output: true
}
