package ratelimit

import (
	"context"
	"log"
	"math"
	"strconv"

	"github.com/ay/go-kit/ginx"
	"github.com/gin-gonic/gin"
)

// safeGo 安全执行函数，recover panic 防止崩溃
func safeGo(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ratelimit] DenyLog goroutine panic: %v", r)
		}
	}()
	fn()
}

// KeyFunc 从请求中提取限流 key
type KeyFunc func(c *gin.Context) string

// GinConfig Gin 中间件配置
type GinConfig struct {
	KeyFunc   KeyFunc                // 提取限流 key（必填）
	Rate      Rate                   // 限流规则（必填）
	ErrorCode int                    // 被限流时的 i18n 错误码（必填）
	Skip      func(*gin.Context) bool // 跳过限流的条件（可选）
	DenyLog   *DenyLogger            // 拒绝日志记录器（可选，传入则自动记录被拒绝的请求）
}

// GinMiddleware 返回 Gin 限流中间件
//
// 请求通过时设置响应头：
//   - X-RateLimit-Limit: 窗口内最大请求数
//   - X-RateLimit-Remaining: 剩余可用次数
//
// 请求被拒绝时额外设置：
//   - Retry-After: 建议重试等待秒数
func GinMiddleware(limiter *Limiter, cfg GinConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.Skip != nil && cfg.Skip(c) {
			c.Next()
			return
		}

		key := cfg.KeyFunc(c)
		if key == "" {
			c.Next()
			return
		}

		result, err := limiter.Allow(c.Request.Context(), key, cfg.Rate)
		if err != nil {
			// Redis 故障时放行，不阻塞业务
			c.Next()
			return
		}

		// 设置限流响应头
		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.Rate.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))

		if !result.Allowed {
			retrySeconds := int(math.Ceil(result.RetryAfter.Seconds()))
			if retrySeconds < 1 {
				retrySeconds = 1
			}
			c.Header("Retry-After", strconv.Itoa(retrySeconds))

			// 记录拒绝日志（异步，不阻塞响应；用 Background 避免请求 ctx 取消）
			if cfg.DenyLog != nil {
				method, path := c.Request.Method, c.Request.URL.Path
				go safeGo(func() {
					cfg.DenyLog.Record(context.Background(), key, method, path)
				})
			}

			_ = c.Error(ginx.NewError(cfg.ErrorCode))
			c.Abort()
			return
		}

		c.Next()
	}
}

// ---- 内置 KeyFunc（独立使用） ----

// KeyByIP 按客户端 IP 限流
//
//	prefix 用于区分不同限流规则，如 "api"、"login"
func KeyByIP(prefix string) KeyFunc {
	return func(c *gin.Context) string {
		return prefix + ":ip:" + c.ClientIP()
	}
}

// KeyByHeader 按请求头值限流
//
//	prefix 用于区分不同限流规则
//	header 是请求头名称，如 "X-Api-Key"
//	头为空时返回空字符串（跳过限流）
func KeyByHeader(prefix, header string) KeyFunc {
	return func(c *gin.Context) string {
		val := c.GetHeader(header)
		if val == "" {
			return ""
		}
		return prefix + ":hdr:" + val
	}
}

// KeyByParam 按路由参数限流
//
//	prefix 用于区分不同限流规则
//	param 是路由参数名，如 "id"
func KeyByParam(prefix, param string) KeyFunc {
	return func(c *gin.Context) string {
		val := c.Param(param)
		if val == "" {
			return ""
		}
		return prefix + ":param:" + val
	}
}

// ---- 维度提取器（仅供 KeyCompose 组合使用） ----

// PartByIP 提取 IP 维度片段
func PartByIP() KeyFunc {
	return func(c *gin.Context) string {
		return "ip:" + c.ClientIP()
	}
}

// PartByHeader 提取请求头维度片段，头为空时返回空字符串（中止组合）
func PartByHeader(header string) KeyFunc {
	return func(c *gin.Context) string {
		val := c.GetHeader(header)
		if val == "" {
			return ""
		}
		return "hdr:" + val
	}
}

// PartByParam 提取路由参数维度片段，值为空时返回空字符串（中止组合）
func PartByParam(param string) KeyFunc {
	return func(c *gin.Context) string {
		val := c.Param(param)
		if val == "" {
			return ""
		}
		return "param:" + val
	}
}

// KeyCompose 组合 prefix + 多个维度提取器生成 key
//
//	任一提取器返回空字符串时跳过限流。
//	例: KeyCompose("login", PartByIP(), PartByHeader("X-User"))
//	生成: "login:ip:1.2.3.4:hdr:alice"
func KeyCompose(prefix string, parts ...KeyFunc) KeyFunc {
	return func(c *gin.Context) string {
		key := prefix
		for _, fn := range parts {
			part := fn(c)
			if part == "" {
				return ""
			}
			key += ":" + part
		}
		return key
	}
}
