//go:build ignore
// +build ignore

package main

import (
	"context"
	"github.com/goforj/scheduler"
	"github.com/redis/go-redis/v9"
	"time"
)

func main() {
	// NewRedisLocker creates a RedisLocker with a client and TTL.
	// The ttl is a lease duration: when it expires, another worker may acquire the
	// same lock key. For long-running jobs, choose ttl >= worst-case runtime plus a
	// safety buffer. If your runtime can exceed ttl, prefer a renewing/heartbeat lock strategy.

	// Example: create a redis-backed locker
	client := redis.NewClient(&redis.Options{}) // replace with your client
	locker := scheduler.NewRedisLocker(client, 10*time.Minute)
	_, _ = locker.Lock(context.Background(), "job")
}
