package auth

import "context"

// Service defines token management operations
type Service interface {
	Create(ctx context.Context, userID int64, scene string) (*Tokens, error)
	Validate(ctx context.Context, accessToken, refreshToken, scene string) (int64, error)
	Refresh(ctx context.Context, accessToken, refreshToken, scene string) (*Tokens, error)
	Delete(ctx context.Context, userID int64, scene string) error
	GetUserOnlineStatus(ctx context.Context, userIDs []int64, scene string) (map[int64]UserOnlineInfo, error)
}
