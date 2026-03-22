package auth

import (
	"context"
	"fmt"
	"testing"

	"github.com/ay/go-kit/i18n"
)

// mockRepo implements Repository for testing
type mockRepo struct {
	createFn   func(ctx context.Context, userID int64, scene string, tokens *Tokens) error
	validateFn func(ctx context.Context, accessToken, refreshToken, scene string) (int64, error)
	refreshFn  func(ctx context.Context, old, oldR, scene string, newT *Tokens) error
	deleteFn   func(ctx context.Context, userID int64, scene string) error
	onlineFn   func(ctx context.Context, ids []int64, scene string) (map[int64]UserOnlineInfo, error)
}

func (m *mockRepo) Create(ctx context.Context, userID int64, scene string, tokens *Tokens) error {
	if m.createFn != nil {
		return m.createFn(ctx, userID, scene, tokens)
	}
	tokens.AccessToken = "access_test"
	tokens.RefreshToken = "refresh_test"
	return nil
}

func (m *mockRepo) Validate(ctx context.Context, at, rt, scene string) (int64, error) {
	if m.validateFn != nil {
		return m.validateFn(ctx, at, rt, scene)
	}
	return 1, nil
}

func (m *mockRepo) Refresh(ctx context.Context, oldA, oldR, scene string, newT *Tokens) error {
	if m.refreshFn != nil {
		return m.refreshFn(ctx, oldA, oldR, scene, newT)
	}
	newT.AccessToken = "new_access"
	newT.RefreshToken = "new_refresh"
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, uid int64, scene string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, uid, scene)
	}
	return nil
}

func (m *mockRepo) GetUserOnlineStatus(ctx context.Context, ids []int64, scene string) (map[int64]UserOnlineInfo, error) {
	if m.onlineFn != nil {
		return m.onlineFn(ctx, ids, scene)
	}
	return map[int64]UserOnlineInfo{}, nil
}

func TestCreate_Success(t *testing.T) {
	svc := NewService(&mockRepo{}, nil)
	tokens, err := svc.Create(context.Background(), 1, "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens.AccessToken == "" {
		t.Fatal("empty access token")
	}
}

func TestCreate_RetryOnConflict(t *testing.T) {
	attempts := 0
	repo := &mockRepo{
		createFn: func(_ context.Context, _ int64, _ string, tokens *Tokens) error {
			attempts++
			if attempts < 3 {
				return fmt.Errorf("token conflict: code=-1")
			}
			tokens.AccessToken = "ok"
			tokens.RefreshToken = "ok"
			return nil
		},
	}
	svc := NewService(repo, nil)
	tokens, err := svc.Create(context.Background(), 1, "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens.AccessToken != "ok" {
		t.Fatal("wrong token")
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestCreate_MaxRetriesExhausted(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, _ int64, _ string, _ *Tokens) error {
			return fmt.Errorf("token conflict: code=-1")
		},
	}
	svc := NewService(repo, nil)
	_, err := svc.Create(context.Background(), 1, "user")
	if err == nil {
		t.Fatal("expected error")
	}
	// Should return CodeTokenCreateFailed
	if bizErr, ok := err.(interface{ GetCode() int }); ok {
		if bizErr.GetCode() != i18n.CodeTokenCreateFailed {
			t.Fatalf("expected code %d, got %d", i18n.CodeTokenCreateFailed, bizErr.GetCode())
		}
	}
}

func TestValidate_Success(t *testing.T) {
	svc := NewService(&mockRepo{}, nil)
	uid, err := svc.Validate(context.Background(), "a", "r", "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uid != 1 {
		t.Fatalf("expected uid 1, got %d", uid)
	}
}

func TestDelete_Success(t *testing.T) {
	svc := NewService(&mockRepo{}, nil)
	err := svc.Delete(context.Background(), 1, "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRefresh_Success(t *testing.T) {
	svc := NewService(&mockRepo{}, nil)
	tokens, err := svc.Refresh(context.Background(), "old_a", "old_r", "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens.AccessToken != "new_access" {
		t.Fatal("wrong access token")
	}
}

func TestIsConflictError(t *testing.T) {
	if isConflictError(nil) {
		t.Fatal("nil should not be conflict")
	}
	if !isConflictError(fmt.Errorf("token conflict: code=-1")) {
		t.Fatal("should detect conflict")
	}
	if isConflictError(fmt.Errorf("some other error")) {
		t.Fatal("should not detect non-conflict")
	}
}
