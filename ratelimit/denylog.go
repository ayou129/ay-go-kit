package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"os"
	"sync/atomic"
	"time"

	"github.com/ay/go-kit/rediscli"
)

const (
	defaultDenyTTL = 90 * 24 * time.Hour // 90 天
)

var (
	denySeq atomic.Int64
	denyPid = os.Getpid()
)

// DenyEntry 单条拒绝日志
type DenyEntry struct {
	Timestamp time.Time
	Method    string
	Path      string
}

// DenyTopEntry 拒绝排行条目
type DenyTopEntry struct {
	Key   string // 限流 key（如 api:ip:1.2.3.4）
	Count int64  // 拒绝次数
}

// DenyLogger 限流拒绝日志（按月 ZSET 存储）
type DenyLogger struct {
	client  *rediscli.Client
	project string
	ttl     time.Duration
}

// NewDenyLogger 创建拒绝日志记录器
//
//	project 为子项目标识（如 "tc"、"flowi"），用于在多项目共享 Redis 时隔离日志。
//	传空字符串时使用默认值 "global"。
//	ttl 为日志保留时长，0 表示使用默认值 90 天。
func NewDenyLogger(client *rediscli.Client, project string, ttl time.Duration) *DenyLogger {
	if project == "" {
		project = "global"
	}
	if ttl <= 0 {
		ttl = defaultDenyTTL
	}
	return &DenyLogger{client: client, project: project, ttl: ttl}
}

// denyKey 月度拒绝日志 key: {project}_ratelimit_deny_{month}_{key}
func denyKey(project, month, key string) string {
	return project + "_ratelimit_deny_" + month + "_" + key
}

// denyTopKey 月度拒绝排行 key: {project}_ratelimit_deny_top_{month}
func denyTopKey(project, month string) string {
	return project + "_ratelimit_deny_top_" + month
}

// currentMonth 返回当前年月
func currentMonth() string {
	return time.Now().Format("2006-01")
}

// Record 记录一次拒绝（由中间件调用）
func (dl *DenyLogger) Record(ctx context.Context, key, method, path string) error {
	month := currentMonth()
	nowMs := time.Now().UnixMilli()
	member := fmt.Sprintf("%d:%d:%d:%s:%s", nowMs, denySeq.Add(1), denyPid, method, path)
	ttlSec := int(dl.ttl.Seconds())

	_, err := dl.client.ExecuteLuaScript(luaDenyRecord,
		[]string{denyKey(dl.project, month, key), denyTopKey(dl.project, month)},
		nowMs, member, key, ttlSec,
	)
	if err != nil {
		return fmt.Errorf("deny log record: %w", err)
	}
	return nil
}

// Count 查询某个 key 在指定月份的拒绝次数
//
//	month 格式: "2026-03"
func (dl *DenyLogger) Count(ctx context.Context, key, month string) (int64, error) {
	raw, err := dl.client.ExecuteLuaScript(
		`return redis.call("ZCARD", KEYS[1])`,
		[]string{denyKey(dl.project, month, key)},
	)
	if err != nil {
		return 0, fmt.Errorf("deny log count: %w", err)
	}
	count, _ := raw.(int64)
	return count, nil
}

// Entries 分页查询某个 key 在指定月份的拒绝记录（按时间倒序）
//
//	month 格式: "2026-03"，offset 从 0 开始，limit 每页条数
func (dl *DenyLogger) Entries(ctx context.Context, key, month string, offset, limit int) ([]DenyEntry, int64, error) {
	raw, err := dl.client.ExecuteLuaScript(luaDenyEntries,
		[]string{denyKey(dl.project, month, key)},
		offset, offset+limit-1,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("deny log entries: %w", err)
	}

	arr, ok := raw.([]any)
	if !ok || len(arr) < 2 {
		return nil, 0, nil
	}

	total, _ := arr[0].(int64)
	items, _ := arr[1].([]any)

	entries := make([]DenyEntry, 0, len(items)/2)
	for i := 0; i+1 < len(items); i += 2 {
		memberStr, _ := items[i].(string)
		entry := parseDenyMember(memberStr)
		// score 作为时间戳备用
		scoreStr, _ := items[i+1].(string)
		if entry.Timestamp.IsZero() {
			var ms int64
			fmt.Sscanf(scoreStr, "%d", &ms)
			entry.Timestamp = time.UnixMilli(ms)
		}
		entries = append(entries, entry)
	}

	return entries, total, nil
}

// Top 查询指定月份拒绝次数排行（降序）
//
//	month 格式: "2026-03"，limit 返回前 N 名
func (dl *DenyLogger) Top(ctx context.Context, month string, limit int) ([]DenyTopEntry, error) {
	raw, err := dl.client.ExecuteLuaScript(luaDenyTop,
		[]string{denyTopKey(dl.project, month)},
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("deny log top: %w", err)
	}

	items, ok := raw.([]any)
	if !ok {
		return nil, nil
	}

	entries := make([]DenyTopEntry, 0, len(items)/2)
	for i := 0; i+1 < len(items); i += 2 {
		key, _ := items[i].(string)
		scoreStr, _ := items[i+1].(string)
		count, _ := strconv.ParseInt(scoreStr, 10, 64)
		entries = append(entries, DenyTopEntry{Key: key, Count: count})
	}
	return entries, nil
}

// Cleanup 清理指定月份的全部拒绝日志
//
//	month 格式: "2026-03"
//	通常不需要手动调用——Redis TTL 会自动过期。
//	此方法用于主动清理（如磁盘告警时）。
func (dl *DenyLogger) Cleanup(ctx context.Context, month string) (int64, error) {
	rdb := dl.client.Redis()
	pattern := dl.project + "_ratelimit_deny_" + month + "_*"

	var deleted int64
	var cursor uint64
	for {
		keys, next, err := rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return deleted, fmt.Errorf("deny log cleanup scan: %w", err)
		}
		if len(keys) > 0 {
			n, err := rdb.Del(ctx, keys...).Result()
			if err != nil {
				return deleted, fmt.Errorf("deny log cleanup del: %w", err)
			}
			deleted += n
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}

	// 清理对应月份的排行 key
	topKey := denyTopKey(dl.project, month)
	n, _ := rdb.Del(ctx, topKey).Result()
	deleted += n

	return deleted, nil
}

// parseDenyMember 解析 member 格式: "1234567890:42:12345:POST:/api/login"
func parseDenyMember(member string) DenyEntry {
	parts := strings.SplitN(member, ":", 5)
	if len(parts) < 5 {
		return DenyEntry{}
	}
	ms, _ := strconv.ParseInt(parts[0], 10, 64)
	return DenyEntry{
		Timestamp: time.UnixMilli(ms),
		Method:    parts[3],
		Path:      parts[4],
	}
}
