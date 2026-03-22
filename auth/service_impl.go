package auth

import (
	"context"

	"github.com/ay/go-kit/ginx"
	"github.com/ay/go-kit/i18n"
)

const maxRetries = 3

// LogFunc is an optional logger for the auth service
type LogFunc func(ctx context.Context, level, format string, args ...any)

type service struct {
	repo  Repository
	logFn LogFunc
}

// NewService creates a new token service
func NewService(repo Repository, logFn LogFunc) Service {
	return &service{repo: repo, logFn: logFn}
}

func (s *service) log(ctx context.Context, level, format string, args ...any) {
	if s.logFn != nil {
		s.logFn(ctx, level, format, args...)
	}
}

func (s *service) Create(ctx context.Context, userID int64, scene string) (*Tokens, error) {
	for i := 0; i < maxRetries; i++ {
		tokens := &Tokens{}
		err := s.repo.Create(ctx, userID, scene, tokens)
		if err == nil {
			return tokens, nil
		}
		if isConflictError(err) {
			s.log(ctx, "warn", "Token create conflict, retrying: userID=%d, scene=%s, attempt=%d", userID, scene, i+1)
			continue
		}
		return nil, err
	}
	s.log(ctx, "error", "Token create failed after max retries: userID=%d, scene=%s", userID, scene)
	return nil, ginx.NewError(i18n.CodeTokenCreateFailed)
}

func (s *service) Validate(ctx context.Context, accessToken, refreshToken, scene string) (int64, error) {
	return s.repo.Validate(ctx, accessToken, refreshToken, scene)
}

func (s *service) Refresh(ctx context.Context, accessToken, refreshToken, scene string) (*Tokens, error) {
	for i := 0; i < maxRetries; i++ {
		tokens := &Tokens{}
		err := s.repo.Refresh(ctx, accessToken, refreshToken, scene, tokens)
		if err == nil {
			return tokens, nil
		}
		if bizErr, ok := err.(ginx.BizError); ok {
			if bizErr.GetCode() == i18n.CodeTokenInvalid {
				return nil, err
			}
		}
		if isConflictError(err) {
			s.log(ctx, "warn", "Token refresh conflict, retrying: scene=%s, attempt=%d", scene, i+1)
			continue
		}
		return nil, err
	}
	s.log(ctx, "error", "Token refresh failed after max retries: scene=%s", scene)
	return nil, ginx.NewError(i18n.CodeTokenRefreshFailed)
}

func (s *service) Delete(ctx context.Context, userID int64, scene string) error {
	return s.repo.Delete(ctx, userID, scene)
}

func (s *service) GetUserOnlineStatus(ctx context.Context, userIDs []int64, scene string) (map[int64]UserOnlineInfo, error) {
	return s.repo.GetUserOnlineStatus(ctx, userIDs, scene)
}

func isConflictError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return msg == "token conflict: code=-1" ||
		msg == "token conflict: code=-2" ||
		msg == "token conflict: code=-3"
}
