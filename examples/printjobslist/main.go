package main

import "github.com/goforj/scheduler"

func main() {
	// PrintJobsList renders and prints the scheduler job table to stdout.

	// Example: print current jobs
	s := scheduler.New()
	defer s.Stop()
	s.EverySecond().Name("heartbeat").Do(func() {})
	s.PrintJobsList()
	// Output:
	// +------------------------------------------------------------------------------------------+
	// | Scheduler Jobs › (1)
	// +-----------+----------+----------+-----------------------+--------------------+-----------+
	// | Name      | Type     | Schedule | Handler               | Next Run           | Tags      |
	// +-----------+----------+----------+-----------------------+--------------------+-----------+
	// | heartbeat | function | every 1s | main.main (anon func) | in 1s Mar 3 2:15AM | env=local |
	// +-----------+----------+----------+-----------------------+--------------------+-----------+
}
