package ratelimit

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ay/go-kit/rediscli"
)

func setupLimiter(t *testing.T) *Limiter {
	t.Helper()
	addr := "127.0.0.1:6379"
	if v := os.Getenv("REDIS_HOST"); v != "" {
		port := os.Getenv("REDIS_PORT")
		if port == "" {
			port = "6379"
		}
		addr = v + ":" + port
	}
	client, err := rediscli.Open(rediscli.Config{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       15, // 测试用 DB
	})
	if err != nil {
		t.Skipf("Redis 不可用，跳过测试: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	return New(client, "test")
}

func TestAllow_Basic(t *testing.T) {
	limiter := setupLimiter(t)
	ctx := context.Background()
	key := "test:allow_basic:" + time.Now().Format("150405.000")

	t.Cleanup(func() { limiter.Reset(ctx, key) })

	rate := Rate{Limit: 3, Window: 10 * time.Second}

	// 前 3 次应该放行
	for i := 0; i < 3; i++ {
		result, err := limiter.Allow(ctx, key, rate)
		if err != nil {
			t.Fatalf("Allow 第 %d 次失败: %v", i+1, err)
		}
		if !result.Allowed {
			t.Fatalf("Allow 第 %d 次应该放行", i+1)
		}
		if result.Remaining != 3-i-1 {
			t.Fatalf("Allow 第 %d 次 Remaining 应为 %d，实际 %d", i+1, 3-i-1, result.Remaining)
		}
		if result.Total != i+1 {
			t.Fatalf("Allow 第 %d 次 Total 应为 %d，实际 %d", i+1, i+1, result.Total)
		}
	}

	// 第 4 次应该被拒绝
	result, err := limiter.Allow(ctx, key, rate)
	if err != nil {
		t.Fatalf("Allow 第 4 次失败: %v", err)
	}
	if result.Allowed {
		t.Fatal("Allow 第 4 次应该被拒绝")
	}
	if result.RetryAfter <= 0 {
		t.Fatal("RetryAfter 应该大于 0")
	}
}

func TestAllow_WindowExpiry(t *testing.T) {
	limiter := setupLimiter(t)
	ctx := context.Background()
	key := "test:allow_expiry:" + time.Now().Format("150405.000")

	t.Cleanup(func() { limiter.Reset(ctx, key) })

	rate := Rate{Limit: 2, Window: 1 * time.Second}

	// 用完 2 次
	for i := 0; i < 2; i++ {
		limiter.Allow(ctx, key, rate)
	}

	// 第 3 次被拒绝
	result, _ := limiter.Allow(ctx, key, rate)
	if result.Allowed {
		t.Fatal("应该被拒绝")
	}

	// 等待窗口过期
	time.Sleep(1100 * time.Millisecond)

	// 窗口重置后应该放行
	result, err := limiter.Allow(ctx, key, rate)
	if err != nil {
		t.Fatalf("窗口过期后 Allow 失败: %v", err)
	}
	if !result.Allowed {
		t.Fatal("窗口过期后应该放行")
	}
}

func TestPeek_DoesNotConsume(t *testing.T) {
	limiter := setupLimiter(t)
	ctx := context.Background()
	key := "test:peek:" + time.Now().Format("150405.000")

	t.Cleanup(func() { limiter.Reset(ctx, key) })

	rate := Rate{Limit: 3, Window: 10 * time.Second}

	// 消耗 1 次
	limiter.Allow(ctx, key, rate)

	// Peek 不应消耗
	result, err := limiter.Peek(ctx, key, rate)
	if err != nil {
		t.Fatalf("Peek 失败: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("Peek Total 应为 1，实际 %d", result.Total)
	}
	if result.Remaining != 2 {
		t.Fatalf("Peek Remaining 应为 2，实际 %d", result.Remaining)
	}

	// 再次 Peek，Total 不变
	result2, _ := limiter.Peek(ctx, key, rate)
	if result2.Total != 1 {
		t.Fatalf("再次 Peek Total 应为 1，实际 %d", result2.Total)
	}
}

func TestEntries(t *testing.T) {
	limiter := setupLimiter(t)
	ctx := context.Background()
	key := "test:entries:" + time.Now().Format("150405.000")

	t.Cleanup(func() { limiter.Reset(ctx, key) })

	rate := Rate{Limit: 10, Window: 10 * time.Second}

	// 写入 3 条
	for i := 0; i < 3; i++ {
		limiter.Allow(ctx, key, rate)
		time.Sleep(5 * time.Millisecond) // 确保时间戳不同
	}

	entries, err := limiter.Entries(ctx, key, rate)
	if err != nil {
		t.Fatalf("Entries 失败: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("Entries 应返回 3 条，实际 %d", len(entries))
	}

	// 时间应该递增
	for i := 1; i < len(entries); i++ {
		if !entries[i].Timestamp.After(entries[i-1].Timestamp) {
			t.Fatalf("Entries 时间戳应递增: %v >= %v", entries[i-1].Timestamp, entries[i].Timestamp)
		}
	}
}

func TestAllow_Concurrent(t *testing.T) {
	limiter := setupLimiter(t)
	ctx := context.Background()
	key := "test:concurrent:" + time.Now().Format("150405.000")

	t.Cleanup(func() { limiter.Reset(ctx, key) })

	rate := Rate{Limit: 50, Window: 10 * time.Second}

	// 100 个 goroutine 同时请求，限制 50
	const goroutines = 100
	results := make(chan Result, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			r, err := limiter.Allow(ctx, key, rate)
			if err != nil {
				t.Errorf("并发 Allow 失败: %v", err)
				results <- Result{}
				return
			}
			results <- r
		}()
	}

	allowed := 0
	denied := 0
	for i := 0; i < goroutines; i++ {
		r := <-results
		if r.Allowed {
			allowed++
		} else {
			denied++
		}
	}

	if allowed != 50 {
		t.Fatalf("并发 50 限制下应放行 50 个，实际 %d（拒绝 %d）", allowed, denied)
	}
}

func TestAllow_LimitOne(t *testing.T) {
	limiter := setupLimiter(t)
	ctx := context.Background()
	key := "test:limit_one:" + time.Now().Format("150405.000")

	t.Cleanup(func() { limiter.Reset(ctx, key) })

	rate := Rate{Limit: 1, Window: 10 * time.Second}

	r1, _ := limiter.Allow(ctx, key, rate)
	if !r1.Allowed {
		t.Fatal("第 1 次应该放行")
	}
	if r1.Remaining != 0 {
		t.Fatalf("Remaining 应为 0，实际 %d", r1.Remaining)
	}

	r2, _ := limiter.Allow(ctx, key, rate)
	if r2.Allowed {
		t.Fatal("第 2 次应该被拒绝")
	}
}

func TestReset(t *testing.T) {
	limiter := setupLimiter(t)
	ctx := context.Background()
	key := "test:reset:" + time.Now().Format("150405.000")

	rate := Rate{Limit: 2, Window: 10 * time.Second}

	// 用完
	limiter.Allow(ctx, key, rate)
	limiter.Allow(ctx, key, rate)

	// 重置
	if err := limiter.Reset(ctx, key); err != nil {
		t.Fatalf("Reset 失败: %v", err)
	}

	// 重置后应该放行
	result, _ := limiter.Allow(ctx, key, rate)
	if !result.Allowed {
		t.Fatal("Reset 后应该放行")
	}
	if result.Total != 1 {
		t.Fatalf("Reset 后 Total 应为 1，实际 %d", result.Total)
	}
}
