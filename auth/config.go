package auth

import (
	"os"
	"strconv"
)

// Config holds token service configuration
type Config struct {
	Prefix        string // Redis key prefix
	AccessExpire  int    // Access token TTL in seconds
	RefreshExpire int    // Refresh token TTL in seconds
}

// DefaultConfig returns config with sensible defaults
func DefaultConfig(prefix string) Config {
	if prefix == "" {
		prefix = "app_token_"
	}
	return Config{
		Prefix:        prefix,
		AccessExpire:  7200,   // 2 hours
		RefreshExpire: 604800, // 7 days
	}
}

// ConfigFromEnv reads token config from environment variables
func ConfigFromEnv() Config {
	prefix := os.Getenv("TOKEN_PREFIX")
	if prefix == "" {
		prefix = "app_token_"
	}
	accessExpire := 7200
	if val := os.Getenv("TOKEN_ACCESS_EXPIRE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			accessExpire = parsed
		}
	}
	refreshExpire := 604800
	if val := os.Getenv("TOKEN_REFRESH_EXPIRE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			refreshExpire = parsed
		}
	}
	return Config{Prefix: prefix, AccessExpire: accessExpire, RefreshExpire: refreshExpire}
}
