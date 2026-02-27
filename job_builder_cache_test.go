package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockCacheLockClient struct {
	tryLockResult bool
	tryLockErr    error
	unlockErr     error
	unlockCalls   int
}

func (m *mockCacheLockClient) TryLock(key string, ttl time.Duration) (bool, error) {
	return m.tryLockResult, m.tryLockErr
}

func (m *mockCacheLockClient) TryLockCtx(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return m.tryLockResult, m.tryLockErr
}

func (m *mockCacheLockClient) Lock(key string, ttl, timeout time.Duration) (bool, error) {
	return m.tryLockResult, m.tryLockErr
}

func (m *mockCacheLockClient) LockCtx(ctx context.Context, key string, ttl, retryInterval time.Duration) (bool, error) {
	return m.tryLockResult, m.tryLockErr
}

func (m *mockCacheLockClient) Unlock(key string) error {
	m.unlockCalls++
	return m.unlockErr
}

func (m *mockCacheLockClient) UnlockCtx(ctx context.Context, key string) error {
	m.unlockCalls++
	return m.unlockErr
}

func TestCacheLockerSuccess(t *testing.T) {
	client := &mockCacheLockClient{tryLockResult: true}
	locker := NewCacheLocker(client, time.Minute)

	lock, err := locker.Lock(context.Background(), "job1")
	require.NoError(t, err)
	require.NotNil(t, lock)
	require.NoError(t, lock.Unlock(context.Background()))
	require.Equal(t, 1, client.unlockCalls)
}

func TestCacheLockerNotAcquired(t *testing.T) {
	client := &mockCacheLockClient{tryLockResult: false}
	locker := NewCacheLocker(client, time.Minute)

	_, err := locker.Lock(context.Background(), "job1")
	require.ErrorIs(t, err, errLockNotAcquired)
}

func TestCacheLockerTryLockError(t *testing.T) {
	client := &mockCacheLockClient{tryLockErr: errors.New("boom")}
	locker := NewCacheLocker(client, time.Minute)

	_, err := locker.Lock(context.Background(), "job1")
	require.EqualError(t, err, "boom")
}

func TestCacheLockerUnlockError(t *testing.T) {
	client := &mockCacheLockClient{
		tryLockResult: true,
		unlockErr:     errors.New("unlock failed"),
	}
	locker := NewCacheLocker(client, time.Minute)

	lock, err := locker.Lock(context.Background(), "job1")
	require.NoError(t, err)
	require.EqualError(t, lock.Unlock(context.Background()), "unlock failed")
	require.Equal(t, 1, client.unlockCalls)
}
