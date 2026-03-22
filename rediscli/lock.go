package rediscli

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// DistributedLock implements a Redis-based distributed lock using SET NX EX
type DistributedLock struct {
	client *redis.Client
	key    string
	value  string
	ttl    time.Duration
}

// NewDistributedLock creates a distributed lock instance
func NewDistributedLock(client *redis.Client, key string, ttl time.Duration) *DistributedLock {
	return &DistributedLock{
		client: client,
		key:    key,
		value:  fmt.Sprintf("%d", time.Now().UnixNano()),
		ttl:    ttl,
	}
}

// Lock acquires the lock, blocking until success or timeout
func (l *DistributedLock) Lock(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		ok, err := l.client.SetNX(ctx, l.key, l.value, l.ttl).Result()
		if err != nil {
			return fmt.Errorf("redis lock: %w", err)
		}
		if ok {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("lock timeout after %v", timeout)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TryLock attempts to acquire the lock without blocking
func (l *DistributedLock) TryLock(ctx context.Context) (bool, error) {
	ok, err := l.client.SetNX(ctx, l.key, l.value, l.ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis lock: %w", err)
	}
	return ok, nil
}

// Unlock releases the lock if owned by this instance (Lua script for atomicity)
func (l *DistributedLock) Unlock(ctx context.Context) error {
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Result()
	if err != nil {
		return fmt.Errorf("unlock: %w", err)
	}
	if result == int64(0) {
		return fmt.Errorf("lock not held by this client")
	}
	return nil
}

// Extend extends the lock TTL if still owned
func (l *DistributedLock) Extend(ctx context.Context, additionalTTL time.Duration) error {
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("EXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`
	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value, int(additionalTTL.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("extend lock: %w", err)
	}
	if result == int64(0) {
		return fmt.Errorf("lock not held by this client")
	}
	return nil
}

// BatchLock atomically acquires multiple locks (all-or-nothing via Lua script)
type BatchLock struct {
	client *redis.Client
	keys   []string
	value  string
	ttl    time.Duration
}

// NewBatchLock creates a batch lock instance
func NewBatchLock(client *redis.Client, keys []string, ttl time.Duration) *BatchLock {
	return &BatchLock{
		client: client,
		keys:   keys,
		value:  fmt.Sprintf("%d", time.Now().UnixNano()),
		ttl:    ttl,
	}
}

// Lock acquires all locks atomically, blocking until success or timeout
func (l *BatchLock) Lock(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	script := `
		local value = ARGV[1]
		local ttl = tonumber(ARGV[2])
		for i, key in ipairs(KEYS) do
			local ok = redis.call("SET", key, value, "NX", "EX", ttl)
			if not ok then
				for j = 1, i-1 do
					redis.call("DEL", KEYS[j])
				end
				return 0
			end
		end
		return 1
	`
	for {
		result, err := l.client.Eval(ctx, script, l.keys, l.value, int(l.ttl.Seconds())).Result()
		if err != nil {
			return fmt.Errorf("batch lock: %w", err)
		}
		if result == int64(1) {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("batch lock timeout after %v", timeout)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// Unlock releases all locks owned by this instance
func (l *BatchLock) Unlock(ctx context.Context) error {
	script := `
		local value = ARGV[1]
		local deleted = 0
		for i, key in ipairs(KEYS) do
			if redis.call("GET", key) == value then
				redis.call("DEL", key)
				deleted = deleted + 1
			end
		end
		return deleted
	`
	result, err := l.client.Eval(ctx, script, l.keys, l.value).Result()
	if err != nil {
		return fmt.Errorf("batch unlock: %w", err)
	}
	if result == int64(0) {
		return fmt.Errorf("no locks held by this client")
	}
	return nil
}
