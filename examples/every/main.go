package main

import "github.com/goforj/scheduler"

func main() {
	// Every starts an interval chain identical to JobBuilder.Every.

	// Example: fluently choose an interval
	scheduler.New().Every(10).Minutes()
}
