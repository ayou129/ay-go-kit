package auth

import "context"

// Repository defines token storage operations (typically backed by Redis)
type Repository interface {
	Create(ctx context.Context, userID int64, scene string, tokens *Tokens) error
	Validate(ctx context.Context, accessToken, refreshToken, scene string) (int64, error)
	Refresh(ctx context.Context, oldAccessToken, oldRefreshToken, scene string, newTokens *Tokens) error
	Delete(ctx context.Context, userID int64, scene string) error
	GetUserOnlineStatus(ctx context.Context, userIDs []int64, scene string) (map[int64]UserOnlineInfo, error)
}
