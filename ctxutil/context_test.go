package ctxutil

import (
	"context"
	"testing"
)

func TestUid(t *testing.T) {
	t.Run("nil context returns 0", func(t *testing.T) {
		if got := GetUid(nil); got != 0 {
			t.Errorf("GetUid(nil) = %d, want 0", got)
		}
	})
	t.Run("missing value returns 0", func(t *testing.T) {
		if got := GetUid(context.Background()); got != 0 {
			t.Errorf("GetUid(bg) = %d, want 0", got)
		}
	})
	t.Run("set-get roundtrip", func(t *testing.T) {
		ctx := WithUid(context.Background(), 42)
		if got := GetUid(ctx); got != 42 {
			t.Errorf("GetUid = %d, want 42", got)
		}
	})
}

func TestTenantID(t *testing.T) {
	t.Run("nil context returns 0", func(t *testing.T) {
		if got := GetTenantID(nil); got != 0 {
			t.Errorf("GetTenantID(nil) = %d, want 0", got)
		}
	})
	t.Run("missing value returns 0", func(t *testing.T) {
		if got := GetTenantID(context.Background()); got != 0 {
			t.Errorf("GetTenantID(bg) = %d, want 0", got)
		}
	})
	t.Run("set-get roundtrip", func(t *testing.T) {
		ctx := WithTenantID(context.Background(), 99)
		if got := GetTenantID(ctx); got != 99 {
			t.Errorf("GetTenantID = %d, want 99", got)
		}
	})
}

func TestTraceID(t *testing.T) {
	t.Run("nil context returns empty", func(t *testing.T) {
		if got := GetTraceID(nil); got != "" {
			t.Errorf("GetTraceID(nil) = %q, want empty", got)
		}
	})
	t.Run("missing value returns empty", func(t *testing.T) {
		if got := GetTraceID(context.Background()); got != "" {
			t.Errorf("GetTraceID(bg) = %q, want empty", got)
		}
	})
	t.Run("set-get roundtrip", func(t *testing.T) {
		ctx := WithTraceID(context.Background(), "trace-abc-123")
		if got := GetTraceID(ctx); got != "trace-abc-123" {
			t.Errorf("GetTraceID = %q, want %q", got, "trace-abc-123")
		}
	})
}

func TestLang(t *testing.T) {
	t.Run("nil context returns zh", func(t *testing.T) {
		if got := GetLang(nil); got != "zh" {
			t.Errorf("GetLang(nil) = %q, want zh", got)
		}
	})
	t.Run("missing value returns zh", func(t *testing.T) {
		if got := GetLang(context.Background()); got != "zh" {
			t.Errorf("GetLang(bg) = %q, want zh", got)
		}
	})
	t.Run("set-get roundtrip", func(t *testing.T) {
		ctx := WithLang(context.Background(), "en")
		if got := GetLang(ctx); got != "en" {
			t.Errorf("GetLang = %q, want en", got)
		}
	})
}

func TestAccessToken(t *testing.T) {
	t.Run("nil context returns empty", func(t *testing.T) {
		if got := GetAccessToken(nil); got != "" {
			t.Errorf("GetAccessToken(nil) = %q, want empty", got)
		}
	})
	t.Run("missing value returns empty", func(t *testing.T) {
		if got := GetAccessToken(context.Background()); got != "" {
			t.Errorf("GetAccessToken(bg) = %q, want empty", got)
		}
	})
	t.Run("set-get roundtrip", func(t *testing.T) {
		ctx := WithAccessToken(context.Background(), "access-tok-xyz")
		if got := GetAccessToken(ctx); got != "access-tok-xyz" {
			t.Errorf("GetAccessToken = %q, want %q", got, "access-tok-xyz")
		}
	})
}

func TestRefreshToken(t *testing.T) {
	t.Run("nil context returns empty", func(t *testing.T) {
		if got := GetRefreshToken(nil); got != "" {
			t.Errorf("GetRefreshToken(nil) = %q, want empty", got)
		}
	})
	t.Run("missing value returns empty", func(t *testing.T) {
		if got := GetRefreshToken(context.Background()); got != "" {
			t.Errorf("GetRefreshToken(bg) = %q, want empty", got)
		}
	})
	t.Run("set-get roundtrip", func(t *testing.T) {
		ctx := WithRefreshToken(context.Background(), "refresh-tok-abc")
		if got := GetRefreshToken(ctx); got != "refresh-tok-abc" {
			t.Errorf("GetRefreshToken = %q, want %q", got, "refresh-tok-abc")
		}
	})
}

func TestMultipleValues(t *testing.T) {
	ctx := context.Background()
	ctx = WithUid(ctx, 1)
	ctx = WithTenantID(ctx, 2)
	ctx = WithTraceID(ctx, "t-123")
	ctx = WithLang(ctx, "en")
	ctx = WithAccessToken(ctx, "at")
	ctx = WithRefreshToken(ctx, "rt")

	if GetUid(ctx) != 1 {
		t.Error("Uid mismatch")
	}
	if GetTenantID(ctx) != 2 {
		t.Error("TenantID mismatch")
	}
	if GetTraceID(ctx) != "t-123" {
		t.Error("TraceID mismatch")
	}
	if GetLang(ctx) != "en" {
		t.Error("Lang mismatch")
	}
	if GetAccessToken(ctx) != "at" {
		t.Error("AccessToken mismatch")
	}
	if GetRefreshToken(ctx) != "rt" {
		t.Error("RefreshToken mismatch")
	}
}
