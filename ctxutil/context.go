package ctxutil

import "context"

type ctxKey string

const (
	ctxKeyUid          ctxKey = "u_id"
	ctxKeyTenantID     ctxKey = "tenant_id"
	ctxKeyTraceID      ctxKey = "trace_id"
	ctxKeyLang         ctxKey = "lang"
	ctxKeyAccessToken  ctxKey = "access_token"
	ctxKeyRefreshToken ctxKey = "refresh_token"
)

// DefaultLang is the fallback language when not set in context.
// Set this at startup (e.g. ctxutil.DefaultLang = "en") to change.
var DefaultLang = "zh"

// Uid

func WithUid(ctx context.Context, uid int64) context.Context {
	return context.WithValue(ctx, ctxKeyUid, uid)
}

func GetUid(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	if uid, ok := ctx.Value(ctxKeyUid).(int64); ok {
		return uid
	}
	return 0
}

// TenantID (generic name, replaces BrandID)

func WithTenantID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, ctxKeyTenantID, id)
}

func GetTenantID(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	if id, ok := ctx.Value(ctxKeyTenantID).(int64); ok {
		return id
	}
	return 0
}

// TraceID

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ctxKeyTraceID, traceID)
}

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(ctxKeyTraceID).(string); ok {
		return traceID
	}
	return ""
}

// Lang

func WithLang(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, ctxKeyLang, lang)
}

func GetLang(ctx context.Context) string {
	if ctx == nil {
		return DefaultLang
	}
	if lang, ok := ctx.Value(ctxKeyLang).(string); ok {
		return lang
	}
	return DefaultLang
}

// AccessToken

func WithAccessToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, ctxKeyAccessToken, token)
}

func GetAccessToken(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if token, ok := ctx.Value(ctxKeyAccessToken).(string); ok {
		return token
	}
	return ""
}

// RefreshToken

func WithRefreshToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, ctxKeyRefreshToken, token)
}

func GetRefreshToken(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if token, ok := ctx.Value(ctxKeyRefreshToken).(string); ok {
		return token
	}
	return ""
}
