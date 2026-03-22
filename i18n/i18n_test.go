package i18n

import (
	"testing"
)

func TestNewCatalog_HasDefaultMessages(t *testing.T) {
	c := NewCatalog(LangZh)

	codes := []int{
		CodeSuccess, CodeInternalError, CodeParamInvalid,
		CodeRouteNotFound, CodeForbidden, CodeRateLimit,
		CodeTokenInvalid, CodeTokenExpired, CodeTokenCreateFailed, CodeTokenRefreshFailed,
		CodeDataNotFound, CodeDataExists, CodeDataConflict,
		CodeWelcome, CodeParamFilterFieldNotAllowed, CodeMaintenance,
	}

	for _, code := range codes {
		msg := c.GetMsg(code, LangZh)
		if msg == "" || msg == "UNKNOWN_ERROR_0" {
			t.Errorf("code %d should have a zh message, got %q", code, msg)
		}
		msg = c.GetMsg(code, LangEn)
		if msg == "" {
			t.Errorf("code %d should have an en message, got %q", code, msg)
		}
	}
}

func TestGetMsg_ZhEn(t *testing.T) {
	c := NewCatalog(LangZh)

	zh := c.GetMsg(CodeSuccess, LangZh)
	if zh != "操作成功" {
		t.Errorf("GetMsg(CodeSuccess, zh) = %q, want 操作成功", zh)
	}

	en := c.GetMsg(CodeSuccess, LangEn)
	if en != "Success" {
		t.Errorf("GetMsg(CodeSuccess, en) = %q, want Success", en)
	}
}

func TestGetMsg_UnknownCode(t *testing.T) {
	c := NewCatalog(LangZh)
	got := c.GetMsg(99999, LangZh)
	want := "UNKNOWN_ERROR_99999"
	if got != want {
		t.Errorf("GetMsg(99999) = %q, want %q", got, want)
	}
}

func TestGetMsg_WithFormatParams(t *testing.T) {
	c := NewCatalog(LangZh)
	c.Register(60001, map[string]string{
		LangZh: "用户 %s 已被禁用",
		LangEn: "User %s has been disabled",
	})

	got := c.GetMsg(60001, LangZh, "Alice")
	want := "用户 Alice 已被禁用"
	if got != want {
		t.Errorf("GetMsg with format = %q, want %q", got, want)
	}

	got = c.GetMsg(60001, LangEn, "Bob")
	want = "User Bob has been disabled"
	if got != want {
		t.Errorf("GetMsg with format (en) = %q, want %q", got, want)
	}
}

func TestGetMsg_WithStringDetailParam(t *testing.T) {
	c := NewCatalog(LangZh)

	// CodeParamInvalid has no % placeholder, so string param appends as detail
	got := c.GetMsg(CodeParamInvalid, LangZh, "email 格式错误")
	want := "参数无效：email 格式错误"
	if got != want {
		t.Errorf("GetMsg with detail = %q, want %q", got, want)
	}
}

func TestGetMsg_WithEmptyStringParam(t *testing.T) {
	c := NewCatalog(LangZh)

	// Empty string detail should not append
	got := c.GetMsg(CodeParamInvalid, LangZh, "")
	want := "参数无效"
	if got != want {
		t.Errorf("GetMsg with empty detail = %q, want %q", got, want)
	}
}

func TestRegister(t *testing.T) {
	c := NewCatalog(LangZh)
	c.Register(70001, map[string]string{
		LangZh: "自定义消息",
		LangEn: "Custom message",
	})

	if got := c.GetMsg(70001, LangZh); got != "自定义消息" {
		t.Errorf("Register zh = %q, want 自定义消息", got)
	}
	if got := c.GetMsg(70001, LangEn); got != "Custom message" {
		t.Errorf("Register en = %q, want Custom message", got)
	}
}

func TestRegisterBatch(t *testing.T) {
	c := NewCatalog(LangZh)
	c.RegisterBatch(map[int]map[string]string{
		80001: {LangZh: "批量A", LangEn: "Batch A"},
		80002: {LangZh: "批量B", LangEn: "Batch B"},
	})

	if got := c.GetMsg(80001, LangZh); got != "批量A" {
		t.Errorf("RegisterBatch 80001 = %q", got)
	}
	if got := c.GetMsg(80002, LangEn); got != "Batch B" {
		t.Errorf("RegisterBatch 80002 = %q", got)
	}
}

func TestGlobal_SetAndGet(t *testing.T) {
	original := Global()
	defer SetGlobal(original) // restore

	c := NewCatalog(LangEn)
	SetGlobal(c)

	if Global() != c {
		t.Error("SetGlobal/Global roundtrip failed")
	}
	if GetDefaultLang() != LangEn {
		t.Errorf("GetDefaultLang = %q, want en", GetDefaultLang())
	}
}

func TestGetLangMsg_UsesGlobal(t *testing.T) {
	got := GetLangMsg(CodeSuccess, LangZh)
	if got != "操作成功" {
		t.Errorf("GetLangMsg = %q, want 操作成功", got)
	}
}

func TestGetMsg_FallbackToDefaultLang(t *testing.T) {
	c := NewCatalog(LangZh)
	c.Register(90001, map[string]string{
		LangZh: "只有中文",
	})

	// Request "en" but only "zh" exists -> falls back to zh
	got := c.GetMsg(90001, LangEn)
	if got != "只有中文" {
		t.Errorf("fallback = %q, want 只有中文", got)
	}
}

func TestGetMsg_EmptyLangUsesDefault(t *testing.T) {
	c := NewCatalog(LangZh)
	got := c.GetMsg(CodeSuccess, "")
	if got != "操作成功" {
		t.Errorf("empty lang = %q, want 操作成功", got)
	}
}

func TestCatalog_GetDefaultLang(t *testing.T) {
	c := NewCatalog(LangEn)
	if got := c.GetDefaultLang(); got != LangEn {
		t.Errorf("GetDefaultLang = %q, want en", got)
	}
}
