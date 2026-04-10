package main

import (
	"github.com/goforj/scheduler/v2"
	"time"
)

func main() {
	// WithNowFunc overrides current time (default: time.Now). Useful for tests.

	// Example: freeze time for predicates
	fixed := func() time.Time { return time.Unix(0, 0) }
	scheduler.New().WithNowFunc(fixed)
}
