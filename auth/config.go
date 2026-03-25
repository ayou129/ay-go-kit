package auth

import (
	"os"
	"strconv"
)

// Config holds token service configuration
type Config struct {
	Project       string // 子项目标识（如 "tc"、"flowi"），Redis key 前缀: {project}_auth_{scene}_{key}
	AccessExpire  int    // Access token TTL in seconds
	RefreshExpire int    // Refresh token TTL in seconds
}

// DefaultConfig returns config with sensible defaults
func DefaultConfig(project string) Config {
	if project == "" {
		project = "global"
	}
	return Config{
		Project:       project,
		AccessExpire:  7200,   // 2 hours
		RefreshExpire: 604800, // 7 days
	}
}

// ConfigFromEnv reads token config from environment variables
//
//	PROJECT_NAME: 子项目标识（如 "tc"），默认 "global"
//	TOKEN_ACCESS_EXPIRE: access token 过期秒数，默认 7200
//	TOKEN_REFRESH_EXPIRE: refresh token 过期秒数，默认 604800
func ConfigFromEnv() Config {
	project := os.Getenv("PROJECT_NAME")
	if project == "" {
		project = "global"
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
	return Config{Project: project, AccessExpire: accessExpire, RefreshExpire: refreshExpire}
}
