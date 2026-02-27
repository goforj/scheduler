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
	// Lock obtains a lock for the job name.

	// Example: acquire a lock
	client := redis.NewClient(&redis.Options{})
	locker := scheduler.NewRedisLocker(client, 10*time.Minute)
	lock, _ := locker.Lock(context.Background(), "job")
	_ = lock.Unlock(context.Background())
}
