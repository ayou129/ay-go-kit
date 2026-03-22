package rediscli

import (
	"testing"
	"time"
)

func TestNewDistributedLock_SetsFields(t *testing.T) {
	lock := NewDistributedLock(nil, "test:lock:key", 30*time.Second)

	if lock.key != "test:lock:key" {
		t.Errorf("expected key 'test:lock:key', got %q", lock.key)
	}
	if lock.ttl != 30*time.Second {
		t.Errorf("expected ttl 30s, got %v", lock.ttl)
	}
	if lock.value == "" {
		t.Error("expected non-empty value")
	}
	if lock.client != nil {
		t.Error("expected nil client")
	}
}

func TestNewBatchLock_SetsFields(t *testing.T) {
	keys := []string{"lock:1", "lock:2", "lock:3"}
	lock := NewBatchLock(nil, keys, 60*time.Second)

	if len(lock.keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(lock.keys))
	}
	if lock.ttl != 60*time.Second {
		t.Errorf("expected ttl 60s, got %v", lock.ttl)
	}
	if lock.value == "" {
		t.Error("expected non-empty value")
	}
	for i, k := range keys {
		if lock.keys[i] != k {
			t.Errorf("key[%d]: expected %q, got %q", i, k, lock.keys[i])
		}
	}
}

func TestDistributedLock_UniqueValues(t *testing.T) {
	lock1 := NewDistributedLock(nil, "key1", 10*time.Second)
	// Small delay to ensure different UnixNano
	time.Sleep(time.Nanosecond)
	lock2 := NewDistributedLock(nil, "key2", 10*time.Second)

	if lock1.value == lock2.value {
		t.Errorf("expected different values, both got %q", lock1.value)
	}
}

func TestBatchLock_UniqueValues(t *testing.T) {
	lock1 := NewBatchLock(nil, []string{"a"}, 10*time.Second)
	time.Sleep(time.Nanosecond)
	lock2 := NewBatchLock(nil, []string{"b"}, 10*time.Second)

	if lock1.value == lock2.value {
		t.Errorf("expected different values, both got %q", lock1.value)
	}
}
