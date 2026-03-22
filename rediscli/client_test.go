package rediscli

import (
	"os"
	"testing"
	"time"
)

func TestConfigFromEnv_Defaults(t *testing.T) {
	// Clear any existing env vars
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_PASSWORD")
	os.Unsetenv("REDIS_DB")
	os.Unsetenv("REDIS_POOL_SIZE")
	os.Unsetenv("REDIS_MIN_IDLE_CONNS")
	os.Unsetenv("REDIS_CONN_MAX_LIFETIME")
	os.Unsetenv("REDIS_CONN_TIMEOUT")
	os.Unsetenv("REDIS_IDLE_TIMEOUT")

	cfg := ConfigFromEnv()

	if cfg.Addr != ":" {
		t.Errorf("expected addr ':', got %q", cfg.Addr)
	}
	if cfg.Password != "" {
		t.Errorf("expected empty password, got %q", cfg.Password)
	}
	if cfg.DB != 0 {
		t.Errorf("expected DB 0, got %d", cfg.DB)
	}
	if cfg.PoolSize != 100 {
		t.Errorf("expected PoolSize 100, got %d", cfg.PoolSize)
	}
	if cfg.MinIdleConns != 10 {
		t.Errorf("expected MinIdleConns 10, got %d", cfg.MinIdleConns)
	}
	if cfg.DialTimeout != 5*time.Second {
		t.Errorf("expected DialTimeout 5s, got %v", cfg.DialTimeout)
	}
	if cfg.ConnMaxIdleTime != 300*time.Second {
		t.Errorf("expected ConnMaxIdleTime 300s, got %v", cfg.ConnMaxIdleTime)
	}
	if cfg.ConnMaxLifetime != 3600*time.Second {
		t.Errorf("expected ConnMaxLifetime 3600s, got %v", cfg.ConnMaxLifetime)
	}
}

func TestConfigFromEnv_CustomValues(t *testing.T) {
	t.Setenv("REDIS_HOST", "myhost")
	t.Setenv("REDIS_PORT", "6380")
	t.Setenv("REDIS_PASSWORD", "secret")
	t.Setenv("REDIS_DB", "3")
	t.Setenv("REDIS_POOL_SIZE", "50")
	t.Setenv("REDIS_MIN_IDLE_CONNS", "5")
	t.Setenv("REDIS_CONN_MAX_LIFETIME", "7200")
	t.Setenv("REDIS_CONN_TIMEOUT", "10")
	t.Setenv("REDIS_IDLE_TIMEOUT", "600")

	cfg := ConfigFromEnv()

	if cfg.Addr != "myhost:6380" {
		t.Errorf("expected addr 'myhost:6380', got %q", cfg.Addr)
	}
	if cfg.Password != "secret" {
		t.Errorf("expected password 'secret', got %q", cfg.Password)
	}
	if cfg.DB != 3 {
		t.Errorf("expected DB 3, got %d", cfg.DB)
	}
	if cfg.PoolSize != 50 {
		t.Errorf("expected PoolSize 50, got %d", cfg.PoolSize)
	}
	if cfg.MinIdleConns != 5 {
		t.Errorf("expected MinIdleConns 5, got %d", cfg.MinIdleConns)
	}
	if cfg.DialTimeout != 10*time.Second {
		t.Errorf("expected DialTimeout 10s, got %v", cfg.DialTimeout)
	}
	if cfg.ConnMaxIdleTime != 600*time.Second {
		t.Errorf("expected ConnMaxIdleTime 600s, got %v", cfg.ConnMaxIdleTime)
	}
	if cfg.ConnMaxLifetime != 7200*time.Second {
		t.Errorf("expected ConnMaxLifetime 7200s, got %v", cfg.ConnMaxLifetime)
	}
}

func TestConfig_StructFields(t *testing.T) {
	cfg := Config{
		Addr:            "localhost:6379",
		Password:        "pass",
		DB:              1,
		PoolSize:        200,
		MinIdleConns:    20,
		DialTimeout:     3 * time.Second,
		ConnMaxIdleTime: 60 * time.Second,
		ConnMaxLifetime: 1800 * time.Second,
	}

	if cfg.Addr != "localhost:6379" {
		t.Errorf("unexpected Addr: %s", cfg.Addr)
	}
	if cfg.Password != "pass" {
		t.Errorf("unexpected Password: %s", cfg.Password)
	}
	if cfg.DB != 1 {
		t.Errorf("unexpected DB: %d", cfg.DB)
	}
	if cfg.PoolSize != 200 {
		t.Errorf("unexpected PoolSize: %d", cfg.PoolSize)
	}
	if cfg.MinIdleConns != 20 {
		t.Errorf("unexpected MinIdleConns: %d", cfg.MinIdleConns)
	}
}
