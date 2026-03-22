package ginx

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/ay/go-kit/i18n"
)

func TestNewError_CreatesAppErrorWithCorrectCodeAndStatus(t *testing.T) {
	err := NewError(i18n.CodeParamInvalid)
	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatal("expected *AppError")
	}
	if appErr.Code != i18n.CodeParamInvalid {
		t.Errorf("expected code %d, got %d", i18n.CodeParamInvalid, appErr.Code)
	}
	if appErr.HttpStatus != http.StatusBadRequest {
		t.Errorf("expected HTTP status %d, got %d", http.StatusBadRequest, appErr.HttpStatus)
	}
}

func TestNewError_WithParams(t *testing.T) {
	err := NewError(i18n.CodeParamInvalid, "name is required")
	appErr := err.(*AppError)
	if len(appErr.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(appErr.Params))
	}
	if appErr.Params[0] != "name is required" {
		t.Errorf("unexpected param: %v", appErr.Params[0])
	}
}

func TestNewInternal_WrapsCauseError(t *testing.T) {
	cause := errors.New("db connection failed")
	err := NewInternal(cause)
	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatal("expected *AppError")
	}
	if appErr.Code != i18n.CodeInternalError {
		t.Errorf("expected code %d, got %d", i18n.CodeInternalError, appErr.Code)
	}
	if appErr.HttpStatus != http.StatusInternalServerError {
		t.Errorf("expected HTTP status %d, got %d", http.StatusInternalServerError, appErr.HttpStatus)
	}
	if appErr.Err != cause {
		t.Error("expected cause to be preserved")
	}
}

func TestNewForbidden_Returns401(t *testing.T) {
	err := NewForbidden()
	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatal("expected *AppError")
	}
	if appErr.Code != i18n.CodeForbidden {
		t.Errorf("expected code %d, got %d", i18n.CodeForbidden, appErr.Code)
	}
	if appErr.HttpStatus != http.StatusUnauthorized {
		t.Errorf("expected HTTP status %d, got %d", http.StatusUnauthorized, appErr.HttpStatus)
	}
}

func TestAppError_Error_ReturnsFormattedString(t *testing.T) {
	err := &AppError{
		Code:       i18n.CodeParamInvalid,
		HttpStatus: http.StatusBadRequest,
	}
	s := err.Error()
	if !strings.Contains(s, "[Code 10002]") {
		t.Errorf("expected error string to contain code, got: %s", s)
	}
}

func TestAppError_Error_WithCause(t *testing.T) {
	cause := errors.New("something broke")
	err := &AppError{
		Code:       i18n.CodeInternalError,
		HttpStatus: http.StatusInternalServerError,
		Err:        cause,
	}
	s := err.Error()
	if !strings.Contains(s, "cause: something broke") {
		t.Errorf("expected error string to contain cause, got: %s", s)
	}
}

func TestAppError_Unwrap_ReturnsCause(t *testing.T) {
	cause := errors.New("root cause")
	err := &AppError{
		Code:       i18n.CodeInternalError,
		HttpStatus: http.StatusInternalServerError,
		Err:        cause,
	}
	if err.Unwrap() != cause {
		t.Error("Unwrap should return the cause error")
	}
}

func TestAppError_Unwrap_ReturnsNilWhenNoCause(t *testing.T) {
	err := &AppError{
		Code:       i18n.CodeParamInvalid,
		HttpStatus: http.StatusBadRequest,
	}
	if err.Unwrap() != nil {
		t.Error("Unwrap should return nil when no cause")
	}
}

func TestGetHttpStatus_MapsCodesCorrectly(t *testing.T) {
	tests := []struct {
		code     int
		expected int
	}{
		{i18n.CodeForbidden, http.StatusUnauthorized},
		{i18n.CodeTokenInvalid, http.StatusUnauthorized},
		{i18n.CodeTokenExpired, http.StatusUnauthorized},
		{i18n.CodeTokenCreateFailed, http.StatusUnauthorized},
		{i18n.CodeTokenRefreshFailed, http.StatusUnauthorized},
		{i18n.CodeRateLimit, http.StatusTooManyRequests},
		{i18n.CodeRouteNotFound, http.StatusNotFound},
		{i18n.CodeInternalError, http.StatusInternalServerError},
		{i18n.CodeParamInvalid, http.StatusBadRequest},
		{i18n.CodeDataNotFound, http.StatusBadRequest},
		{99999, http.StatusBadRequest}, // unknown code defaults to 400
	}
	for _, tt := range tests {
		got := getHttpStatus(tt.code)
		if got != tt.expected {
			t.Errorf("getHttpStatus(%d) = %d, want %d", tt.code, got, tt.expected)
		}
	}
}

func TestNewRateLimit(t *testing.T) {
	err := NewRateLimit()
	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatal("expected *AppError")
	}
	if appErr.HttpStatus != http.StatusTooManyRequests {
		t.Errorf("expected HTTP status %d, got %d", http.StatusTooManyRequests, appErr.HttpStatus)
	}
}

func TestNewErrorWithCause(t *testing.T) {
	cause := errors.New("timeout")
	err := NewErrorWithCause(i18n.CodeDataNotFound, cause, "user")
	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatal("expected *AppError")
	}
	if appErr.Code != i18n.CodeDataNotFound {
		t.Errorf("expected code %d, got %d", i18n.CodeDataNotFound, appErr.Code)
	}
	if !errors.Is(err, cause) {
		t.Error("expected errors.Is to find cause")
	}
	if len(appErr.Params) != 1 || appErr.Params[0] != "user" {
		t.Errorf("unexpected params: %v", appErr.Params)
	}
}

func TestAppError_ImplementsBizError(t *testing.T) {
	var _ BizError = &AppError{}
}
