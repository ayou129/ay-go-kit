package ratelimit

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ay/go-kit/rediscli"
)

func setupDenyLogger(t *testing.T) (*DenyLogger, *Limiter) {
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
		DB:       15,
	})
	if err != nil {
		t.Skipf("Redis 不可用，跳过测试: %v", err)
	}
	t.Cleanup(func() { client.Close() })

	dl := NewDenyLogger(client, "test", 0)
	limiter := New(client, "test")
	return dl, limiter
}

func TestDenyLogger_Record_And_Count(t *testing.T) {
	dl, _ := setupDenyLogger(t)
	ctx := context.Background()
	key := "test:deny:count:" + time.Now().Format("150405.000")
	month := currentMonth()

	t.Cleanup(func() { dl.Cleanup(ctx, month) })

	// 记录 3 次拒绝
	for i := 0; i < 3; i++ {
		if err := dl.Record(ctx, key, "POST", "/api/login"); err != nil {
			t.Fatalf("Record 第 %d 次失败: %v", i+1, err)
		}
		time.Sleep(2 * time.Millisecond) // 确保时间戳不同
	}

	count, err := dl.Count(ctx, key, month)
	if err != nil {
		t.Fatalf("Count 失败: %v", err)
	}
	if count != 3 {
		t.Fatalf("Count 应为 3，实际 %d", count)
	}
}

func TestDenyLogger_Entries_Pagination(t *testing.T) {
	dl, _ := setupDenyLogger(t)
	ctx := context.Background()
	key := "test:deny:entries:" + time.Now().Format("150405.000")
	month := currentMonth()

	t.Cleanup(func() { dl.Cleanup(ctx, month) })

	// 写入 5 条不同路径
	paths := []string{"/api/a", "/api/b", "/api/c", "/api/d", "/api/e"}
	for _, p := range paths {
		dl.Record(ctx, key, "GET", p)
		time.Sleep(2 * time.Millisecond)
	}

	// 第 1 页 (offset=0, limit=3)
	entries, total, err := dl.Entries(ctx, key, month, 0, 3)
	if err != nil {
		t.Fatalf("Entries 失败: %v", err)
	}
	if total != 5 {
		t.Fatalf("total 应为 5，实际 %d", total)
	}
	if len(entries) != 3 {
		t.Fatalf("第 1 页应返回 3 条，实际 %d", len(entries))
	}

	// 倒序：最新的在前
	if entries[0].Path != "/api/e" {
		t.Fatalf("第 1 条应为 /api/e，实际 %s", entries[0].Path)
	}
	if entries[0].Method != "GET" {
		t.Fatalf("Method 应为 GET，实际 %s", entries[0].Method)
	}

	// 第 2 页 (offset=3, limit=3)
	entries2, _, _ := dl.Entries(ctx, key, month, 3, 3)
	if len(entries2) != 2 {
		t.Fatalf("第 2 页应返回 2 条，实际 %d", len(entries2))
	}
}

func TestDenyLogger_Top(t *testing.T) {
	dl, _ := setupDenyLogger(t)
	ctx := context.Background()
	month := currentMonth()

	t.Cleanup(func() { dl.Cleanup(ctx, month) })

	// 模拟不同 key 的拒绝次数
	keyA := "test:deny:top:a:" + time.Now().Format("150405.000")
	keyB := "test:deny:top:b:" + time.Now().Format("150405.000")
	keyC := "test:deny:top:c:" + time.Now().Format("150405.000")

	for i := 0; i < 5; i++ {
		dl.Record(ctx, keyA, "POST", "/api/login")
	}
	for i := 0; i < 2; i++ {
		dl.Record(ctx, keyB, "GET", "/api/data")
	}
	dl.Record(ctx, keyC, "GET", "/api/status")

	top, err := dl.Top(ctx, month, 10)
	if err != nil {
		t.Fatalf("Top 失败: %v", err)
	}
	if len(top) < 3 {
		t.Fatalf("Top 应至少有 3 条，实际 %d", len(top))
	}

	// 第 1 名应是 keyA（5 次）
	found := false
	for _, entry := range top {
		if entry.Key == keyA {
			found = true
			if entry.Count != 5 {
				t.Fatalf("keyA 拒绝次数应为 5，实际 %d", entry.Count)
			}
			break
		}
	}
	if !found {
		t.Fatal("Top 中未找到 keyA")
	}
}

func TestDenyLogger_Top_Order(t *testing.T) {
	dl, _ := setupDenyLogger(t)
	ctx := context.Background()
	month := currentMonth()

	t.Cleanup(func() { dl.Cleanup(ctx, month) })

	keyA := "test:deny:order:a:" + time.Now().Format("150405.000")
	keyB := "test:deny:order:b:" + time.Now().Format("150405.000")
	keyC := "test:deny:order:c:" + time.Now().Format("150405.000")

	// A=10, B=3, C=7
	for i := 0; i < 10; i++ {
		dl.Record(ctx, keyA, "POST", "/a")
	}
	for i := 0; i < 3; i++ {
		dl.Record(ctx, keyB, "POST", "/b")
	}
	for i := 0; i < 7; i++ {
		dl.Record(ctx, keyC, "POST", "/c")
	}

	top, err := dl.Top(ctx, month, 3)
	if err != nil {
		t.Fatalf("Top 失败: %v", err)
	}
	if len(top) < 3 {
		t.Fatalf("Top 应返回 3 条，实际 %d", len(top))
	}

	// 验证降序：A(10) > C(7) > B(3)
	if top[0].Key != keyA || top[0].Count != 10 {
		t.Fatalf("第 1 名应为 keyA(10)，实际 %s(%d)", top[0].Key, top[0].Count)
	}
	if top[1].Key != keyC || top[1].Count != 7 {
		t.Fatalf("第 2 名应为 keyC(7)，实际 %s(%d)", top[1].Key, top[1].Count)
	}
	if top[2].Key != keyB || top[2].Count != 3 {
		t.Fatalf("第 3 名应为 keyB(3)，实际 %s(%d)", top[2].Key, top[2].Count)
	}
}

func TestDenyLogger_NoDuplicateLoss(t *testing.T) {
	dl, _ := setupDenyLogger(t)
	ctx := context.Background()
	key := "test:deny:nodup:" + time.Now().Format("150405.000")
	month := currentMonth()

	t.Cleanup(func() { dl.Cleanup(ctx, month) })

	// 快速连续写入 20 条（同毫秒可能）
	for i := 0; i < 20; i++ {
		dl.Record(ctx, key, "POST", "/api/login")
	}

	count, _ := dl.Count(ctx, key, month)
	if count != 20 {
		t.Fatalf("20 条记录应全部保留，实际 %d（丢失 %d）", count, 20-count)
	}
}

func TestDenyLogger_Cleanup(t *testing.T) {
	dl, _ := setupDenyLogger(t)
	ctx := context.Background()
	key := "test:deny:cleanup:" + time.Now().Format("150405.000")
	month := currentMonth()

	// 写入数据
	dl.Record(ctx, key, "POST", "/api/test")

	count, _ := dl.Count(ctx, key, month)
	if count == 0 {
		t.Fatal("写入后 count 应 > 0")
	}

	// 清理
	deleted, err := dl.Cleanup(ctx, month)
	if err != nil {
		t.Fatalf("Cleanup 失败: %v", err)
	}
	if deleted == 0 {
		t.Fatal("Cleanup 应删除至少 1 个 key")
	}

	// 清理后应为空
	count2, _ := dl.Count(ctx, key, month)
	if count2 != 0 {
		t.Fatalf("Cleanup 后 count 应为 0，实际 %d", count2)
	}
}

func TestDenyLogger_TTL_IsSet(t *testing.T) {
	dl, _ := setupDenyLogger(t)
	ctx := context.Background()
	key := "test:deny:ttl:" + time.Now().Format("150405.000")
	month := currentMonth()

	t.Cleanup(func() { dl.Cleanup(ctx, month) })

	dl.Record(ctx, key, "GET", "/api/test")

	// 验证 key 有 TTL
	rdb := dl.client.Redis()
	ttl, err := rdb.TTL(ctx, denyKey("test", month, key)).Result()
	if err != nil {
		t.Fatalf("TTL 查询失败: %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("key 应有 TTL，实际 %v", ttl)
	}
	// TTL 应接近 90 天（允许误差）
	if ttl < 89*24*time.Hour {
		t.Fatalf("TTL 应接近 90 天，实际 %v", ttl)
	}
}
