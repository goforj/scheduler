//go:build ignore
// +build ignore

package main

import (
	"context"
	"github.com/goforj/cache"
	"github.com/goforj/cache/driver/rediscache"
	"github.com/goforj/scheduler"
	"time"
)

func main() {
	// NewCacheLocker creates a CacheLocker with a cache lock client and TTL.
	// The ttl is a lease duration: when it expires, another worker may acquire the
	// same lock key. For long-running jobs, choose ttl >= worst-case runtime plus a
	// safety buffer. If your runtime can exceed ttl, prefer a renewing/heartbeat lock strategy.

	// Example: use an in-memory cache driver
	client := cache.NewCache(cache.NewMemoryStore(context.Background()))
	locker := scheduler.NewCacheLocker(client, 10*time.Minute)
	_, _ = locker.Lock(context.Background(), "job")

	// Example: use the Redis cache driver
	redisStore := rediscache.New(rediscache.Config{
		Addr: "127.0.0.1:6379",
	})
	redisClient := cache.NewCache(redisStore)
	redisLocker := scheduler.NewCacheLocker(redisClient, 10*time.Minute)
	_, _ = redisLocker.Lock(context.Background(), "job")
}
