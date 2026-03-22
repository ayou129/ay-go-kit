package rediscli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis connection configuration
type Config struct {
	Addr            string
	Password        string
	DB              int
	PoolSize        int
	MinIdleConns    int
	DialTimeout     time.Duration
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
}

// ConfigFromEnv reads Redis config from environment variables
func ConfigFromEnv() Config {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	poolSize, _ := strconv.Atoi(os.Getenv("REDIS_POOL_SIZE"))
	if poolSize == 0 {
		poolSize = 100
	}
	minIdleConns, _ := strconv.Atoi(os.Getenv("REDIS_MIN_IDLE_CONNS"))
	if minIdleConns == 0 {
		minIdleConns = 10
	}
	connMaxLifetime, _ := strconv.Atoi(os.Getenv("REDIS_CONN_MAX_LIFETIME"))
	if connMaxLifetime == 0 {
		connMaxLifetime = 3600
	}
	connTimeout, _ := strconv.Atoi(os.Getenv("REDIS_CONN_TIMEOUT"))
	if connTimeout == 0 {
		connTimeout = 5
	}
	idleTimeout, _ := strconv.Atoi(os.Getenv("REDIS_IDLE_TIMEOUT"))
	if idleTimeout == 0 {
		idleTimeout = 300
	}

	return Config{
		Addr:            fmt.Sprintf("%s:%s", host, port),
		Password:        password,
		DB:              db,
		PoolSize:        poolSize,
		MinIdleConns:    minIdleConns,
		DialTimeout:     time.Duration(connTimeout) * time.Second,
		ConnMaxIdleTime: time.Duration(idleTimeout) * time.Second,
		ConnMaxLifetime: time.Duration(connMaxLifetime) * time.Second,
	}
}

// Client wraps a Redis client with Lua script SHA caching
type Client struct {
	rdb            *redis.Client
	scriptShaCache sync.Map
}

// Open creates a new Redis client and pings to verify connection
func Open(cfg Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:            cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		DialTimeout:     cfg.DialTimeout,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	if c.rdb != nil {
		return c.rdb.Close()
	}
	return nil
}

// Redis returns the underlying go-redis client
func (c *Client) Redis() *redis.Client {
	return c.rdb
}

// ExecuteLuaScript executes a Lua script with SHA caching
func (c *Client) ExecuteLuaScript(script string, keys []string, args ...any) (any, error) {
	ctx := context.Background()

	if v, ok := c.scriptShaCache.Load(script); ok {
		sha := v.(string)
		res, err := c.rdb.EvalSha(ctx, sha, keys, args...).Result()
		if err == nil {
			return res, nil
		}
		if !strings.Contains(strings.ToUpper(err.Error()), "NOSCRIPT") {
			return nil, err
		}
		c.scriptShaCache.Delete(script)
	}

	sha, err := c.rdb.ScriptLoad(ctx, script).Result()
	if err != nil {
		return nil, err
	}
	c.scriptShaCache.Store(script, sha)

	return c.rdb.EvalSha(ctx, sha, keys, args...).Result()
}
