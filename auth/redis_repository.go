package auth

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"

	"github.com/ay/go-kit/ginx"
	"github.com/ay/go-kit/i18n"
	"github.com/ay/go-kit/rediscli"
	"github.com/ay/go-kit/token"
)

//go:embed lua/token_create.lua
var tokenCreateLua string

//go:embed lua/token_refresh.lua
var tokenRefreshLua string

//go:embed lua/token_validate.lua
var tokenValidateLua string

//go:embed lua/token_delete.lua
var tokenDeleteLua string

type redisRepository struct {
	cfg    Config
	client *rediscli.Client
	logFn  LogFunc
}

// NewRedisRepository creates a Redis+Lua backed token repository
func NewRedisRepository(cfg Config, client *rediscli.Client, logFn LogFunc) Repository {
	return &redisRepository{cfg: cfg, client: client, logFn: logFn}
}

func (r *redisRepository) log(ctx context.Context, level, format string, args ...any) {
	if r.logFn != nil {
		r.logFn(ctx, level, format, args...)
	}
}

func (r *redisRepository) Create(ctx context.Context, userID int64, scene string, tokens *Tokens) error {
	sessionID := token.GenerateToken()
	tokens.AccessToken = token.GenerateToken()
	tokens.RefreshToken = token.GenerateToken()

	keys := []string{fmt.Sprintf("%d", userID), sessionID, tokens.AccessToken, tokens.RefreshToken}
	args := []any{r.cfg.Project, scene, r.cfg.AccessExpire, r.cfg.RefreshExpire}

	result, err := r.client.ExecuteLuaScript(tokenCreateLua, keys, args...)
	if err != nil {
		r.log(ctx, "error", "Token create lua error: %v", err)
		return ginx.NewInternal(err)
	}

	code, err := parseResultCode(result)
	if err != nil {
		r.log(ctx, "error", "Token create parse error: %v", err)
		return ginx.NewInternal(err)
	}

	if code == 1 {
		return nil
	}
	if code >= -3 && code <= -1 {
		return fmt.Errorf("token conflict: code=%d", code)
	}

	r.log(ctx, "error", "Token create unexpected code: %d", code)
	return ginx.NewInternal(fmt.Errorf("unexpected code: %d", code))
}

func (r *redisRepository) Validate(ctx context.Context, accessToken, refreshToken, scene string) (int64, error) {
	keys := []string{accessToken, refreshToken}
	args := []any{r.cfg.Project, scene, r.cfg.AccessExpire}

	result, err := r.client.ExecuteLuaScript(tokenValidateLua, keys, args...)
	if err != nil {
		r.log(ctx, "error", "Token validate lua error: %v", err)
		return 0, ginx.NewInternal(err)
	}

	resultArray, ok := result.([]any)
	if !ok || len(resultArray) == 0 {
		r.log(ctx, "error", "Token validate invalid result: %v", result)
		return 0, ginx.NewInternal(fmt.Errorf("invalid result format"))
	}

	if len(resultArray) == 1 {
		if code, ok := resultArray[0].(int64); ok {
			if code == -1 || code == -2 {
				return 0, ginx.NewError(i18n.CodeTokenInvalid)
			}
		}
	}

	sessionInfo := make(map[string]string)
	for idx := 0; idx < len(resultArray); idx += 2 {
		if idx+1 < len(resultArray) {
			key, ok1 := resultArray[idx].(string)
			val, ok2 := resultArray[idx+1].(string)
			if ok1 && ok2 {
				sessionInfo[key] = val
			}
		}
	}

	userIDStr, exists := sessionInfo["user_id"]
	if !exists {
		r.log(ctx, "error", "Token validate missing user_id in session: %v", sessionInfo)
		return 0, ginx.NewInternal(fmt.Errorf("user_id not found in session"))
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		r.log(ctx, "error", "Token validate user_id parse error: %s", userIDStr)
		return 0, ginx.NewInternal(err)
	}

	return userID, nil
}

func (r *redisRepository) Refresh(ctx context.Context, oldAccessToken, oldRefreshToken, scene string, newTokens *Tokens) error {
	newSessionID := token.GenerateToken()
	newTokens.AccessToken = token.GenerateToken()
	newTokens.RefreshToken = token.GenerateToken()

	keys := []string{oldAccessToken, oldRefreshToken, newSessionID, newTokens.AccessToken, newTokens.RefreshToken}
	args := []any{r.cfg.Project, scene, r.cfg.AccessExpire, r.cfg.RefreshExpire}

	result, err := r.client.ExecuteLuaScript(tokenRefreshLua, keys, args...)
	if err != nil {
		r.log(ctx, "error", "Token refresh lua error: %v", err)
		return ginx.NewInternal(err)
	}

	code, err := parseResultCode(result)
	if err != nil {
		r.log(ctx, "error", "Token refresh parse error: %v", err)
		return ginx.NewInternal(err)
	}

	if code == 1 {
		return nil
	}
	if code >= -3 && code <= -1 {
		return fmt.Errorf("token conflict: code=%d", code)
	}
	if code == -4 || code == -5 {
		return ginx.NewError(i18n.CodeTokenInvalid)
	}

	r.log(ctx, "error", "Token refresh unexpected code: %d", code)
	return ginx.NewInternal(fmt.Errorf("unexpected code: %d", code))
}

func (r *redisRepository) Delete(ctx context.Context, userID int64, scene string) error {
	keys := []string{fmt.Sprintf("%d", userID)}
	args := []any{r.cfg.Project, scene}

	result, err := r.client.ExecuteLuaScript(tokenDeleteLua, keys, args...)
	if err != nil {
		r.log(ctx, "error", "Token delete lua error: %v", err)
		return ginx.NewInternal(err)
	}

	code, err := parseResultCode(result)
	if err != nil {
		r.log(ctx, "error", "Token delete parse error: %v", err)
		return ginx.NewInternal(err)
	}

	if code == 1 {
		return nil
	}

	r.log(ctx, "error", "Token delete unexpected code: %d", code)
	return ginx.NewInternal(fmt.Errorf("unexpected code: %d", code))
}

func (r *redisRepository) GetUserOnlineStatus(ctx context.Context, userIDs []int64, scene string) (map[int64]UserOnlineInfo, error) {
	prefix := r.cfg.Project + "_auth_" + scene + "_"
	rdb := r.client.Redis()
	bgCtx := context.Background()
	result := make(map[int64]UserOnlineInfo, len(userIDs))

	for _, uid := range userIDs {
		info := UserOnlineInfo{}

		sessionID, err := rdb.Get(bgCtx, prefix+"user:"+fmt.Sprintf("%d", uid)).Result()
		if err != nil || sessionID == "" {
			result[uid] = info
			continue
		}

		sessionData, err := rdb.HGetAll(bgCtx, prefix+"token_session:"+sessionID).Result()
		if err != nil || len(sessionData) == 0 {
			result[uid] = info
			continue
		}

		if lastAccess, ok := sessionData["last_access"]; ok {
			if ts, parseErr := strconv.ParseInt(lastAccess, 10, 64); parseErr == nil {
				info.LastAccess = ts
			}
		}

		if at, ok := sessionData["access_token"]; ok {
			exists, existsErr := rdb.Exists(bgCtx, prefix+"access_token:"+at).Result()
			if existsErr == nil && exists > 0 {
				info.IsOnline = true
			}
		}

		result[uid] = info
	}

	return result, nil
}

func parseResultCode(result any) (int64, error) {
	resultArray, ok := result.([]any)
	if !ok || len(resultArray) == 0 {
		return 0, fmt.Errorf("invalid result format")
	}
	code, ok := resultArray[0].(int64)
	if !ok {
		return 0, fmt.Errorf("invalid code type")
	}
	return code, nil
}
