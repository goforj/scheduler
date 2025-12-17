//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"github.com/go-co-op/gocron/v2"
	"github.com/goforj/scheduler"
)

func main() {
	// Job returns the last scheduled gocron.Job instance, if available.

	// Example: capture the last job handle
	s, _ := gocron.NewScheduler()
	s.Start()
	defer s.Shutdown()

	b := scheduler.NewJobBuilder(s).EverySecond().Do(func() {})
	fmt.Println(b.Job() != nil)
}
