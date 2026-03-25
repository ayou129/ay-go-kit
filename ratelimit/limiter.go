package ratelimit

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/ay/go-kit/rediscli"
)

// seq 全局递增计数器，pid 进程标识——两者组合确保多进程同毫秒 member 不重复
var (
	seq atomic.Int64
	pid = os.Getpid()
)

// Rate 定义限流规则
type Rate struct {
	Limit  int           // 窗口内最大请求数
	Window time.Duration // 滑动窗口大小
}

// Result 限流判定结果
type Result struct {
	Allowed    bool          // 是否放行
	Total      int           // 当前窗口内已有请求数
	Remaining  int           // 剩余可用次数
	RetryAfter time.Duration // 被拒绝时，多久后可重试
}

// Entry 单条限流日志
type Entry struct {
	Timestamp time.Time // 请求时间
	ID        string    // 唯一标识（时间戳+随机）
}

// Limiter 基于 Redis ZSET 的滑动窗口限流器
type Limiter struct {
	client  *rediscli.Client
	project string // 子项目标识，用于多项目共享 Redis 时隔离 key
}

// New 创建限流器
//
//	project 为子项目标识（如 "tc"、"flowi"），用于在多项目共享 Redis 时隔离 key。
//	传空字符串时使用默认值 "global"。
//	最终 Redis key 格式: {project}_ratelimit_{业务key}
func New(client *rediscli.Client, project string) *Limiter {
	if project == "" {
		project = "global"
	}
	return &Limiter{client: client, project: project}
}

// redisKey 生成 Redis key: {project}_ratelimit_{key}
func (l *Limiter) redisKey(key string) string {
	return l.project + "_ratelimit_" + key
}

// Allow 判定并记录一次请求（原子操作）
//
// 内部 Lua 脚本在单次调用中完成：清理过期条目 → 计数 → 判定 → 写入
func (l *Limiter) Allow(ctx context.Context, key string, rate Rate) (Result, error) {
	nowMs := time.Now().UnixMilli()
	windowMs := rate.Window.Milliseconds()
	member := fmt.Sprintf("%d:%d:%d", nowMs, seq.Add(1), pid)

	raw, err := l.client.ExecuteLuaScript(luaAllow,
		[]string{l.redisKey(key)},
		windowMs, nowMs, rate.Limit, member,
	)
	if err != nil {
		return Result{}, fmt.Errorf("ratelimit allow: %w", err)
	}

	return parseResult(raw, rate.Limit)
}

// Peek 读取当前限流状态（不计入请求，但会清理过期条目以保证计数精确）
func (l *Limiter) Peek(ctx context.Context, key string, rate Rate) (Result, error) {
	nowMs := time.Now().UnixMilli()
	windowMs := rate.Window.Milliseconds()

	raw, err := l.client.ExecuteLuaScript(luaPeek,
		[]string{l.redisKey(key)},
		windowMs, nowMs, rate.Limit,
	)
	if err != nil {
		return Result{}, fmt.Errorf("ratelimit peek: %w", err)
	}

	return parseResult(raw, rate.Limit)
}

// Entries 读取当前窗口内的所有请求时间戳（用于监控/调试）
func (l *Limiter) Entries(ctx context.Context, key string, rate Rate) ([]Entry, error) {
	nowMs := time.Now().UnixMilli()
	windowMs := rate.Window.Milliseconds()

	raw, err := l.client.ExecuteLuaScript(luaEntries,
		[]string{l.redisKey(key)},
		windowMs, nowMs,
	)
	if err != nil {
		return nil, fmt.Errorf("ratelimit entries: %w", err)
	}

	items, ok := raw.([]any)
	if !ok {
		return nil, nil
	}

	entries := make([]Entry, 0, len(items)/2)
	for i := 0; i+1 < len(items); i += 2 {
		id, _ := items[i].(string)
		scoreStr, _ := items[i+1].(string)
		var scoreMs int64
		fmt.Sscanf(scoreStr, "%d", &scoreMs)
		entries = append(entries, Entry{
			Timestamp: time.UnixMilli(scoreMs),
			ID:        id,
		})
	}
	return entries, nil
}

// Reset 清除指定 key 的限流数据
func (l *Limiter) Reset(ctx context.Context, key string) error {
	_, err := l.client.ExecuteLuaScript(
		`redis.call("DEL", KEYS[1]); return 1`,
		[]string{l.redisKey(key)},
	)
	if err != nil {
		return fmt.Errorf("ratelimit reset: %w", err)
	}
	return nil
}

// parseResult 解析 Lua 脚本返回值
func parseResult(raw any, limit int) (Result, error) {
	arr, ok := raw.([]any)
	if !ok || len(arr) < 4 {
		return Result{}, fmt.Errorf("unexpected lua result: %v", raw)
	}

	allowed, _ := arr[0].(int64)
	remaining, _ := arr[1].(int64)
	retryMs, _ := arr[2].(int64)
	total, _ := arr[3].(int64)

	return Result{
		Allowed:    allowed == 1,
		Total:      int(total),
		Remaining:  int(remaining),
		RetryAfter: time.Duration(retryMs) * time.Millisecond,
	}, nil
}
