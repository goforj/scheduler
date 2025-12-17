//go:build ignore
// +build ignore

package main

import (
	"github.com/goforj/scheduler"
	"time"
)

func main() {
	// WithNowFunc overrides current time (default: time.Now). Useful for tests.

	// Example: freeze time for predicates
	fixed := func() time.Time { return time.Unix(0, 0) }
	scheduler.NewJobBuilder(nil).WithNowFunc(fixed)
}
