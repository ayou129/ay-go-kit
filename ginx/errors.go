package ginx

import (
	"fmt"
	"net/http"

	"github.com/ay/go-kit/i18n"
	"github.com/gin-gonic/gin"
)

// BizError is the business error interface
type BizError interface {
	error
	GetCode() int
	GetHttpStatus() int
	GetMessage(lang string) string
	WriteResponse(c *gin.Context)
}

// AppError implements BizError
type AppError struct {
	Code       int
	HttpStatus int
	Params     []any
	Err        error
}

func (e *AppError) Error() string {
	msg := i18n.GetLangMsg(e.Code, i18n.GetDefaultLang(), e.Params...)
	if e.Err != nil {
		return fmt.Sprintf("[Code %d] %s (cause: %v)", e.Code, msg, e.Err)
	}
	return fmt.Sprintf("[Code %d] %s", e.Code, msg)
}

func (e *AppError) GetCode() int      { return e.Code }
func (e *AppError) GetHttpStatus() int { return e.HttpStatus }

func (e *AppError) GetMessage(lang string) string {
	return i18n.GetLangMsg(e.Code, lang, e.Params...)
}

func (e *AppError) WriteResponse(c *gin.Context) {
	lang := getLang(c)
	c.JSON(e.HttpStatus, gin.H{"code": e.Code, "msg": e.GetMessage(lang), "data": nil})
	c.Abort()
}

func (e *AppError) Unwrap() error { return e.Err }

func getHttpStatus(code int) int {
	switch code {
	case i18n.CodeForbidden, i18n.CodeTokenInvalid, i18n.CodeTokenExpired,
		i18n.CodeTokenCreateFailed, i18n.CodeTokenRefreshFailed:
		return http.StatusUnauthorized
	case i18n.CodeRateLimit:
		return http.StatusTooManyRequests
	case i18n.CodeRouteNotFound:
		return http.StatusNotFound
	case i18n.CodeInternalError:
		return http.StatusInternalServerError
	case i18n.CodeMaintenance:
		return http.StatusServiceUnavailable
	case i18n.CodeDataConflict:
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}

// NewError creates a business error
func NewError(code int, params ...any) error {
	return &AppError{Code: code, HttpStatus: getHttpStatus(code), Params: params}
}

// NewErrorWithCause creates a business error with underlying cause
func NewErrorWithCause(code int, cause error, params ...any) error {
	return &AppError{Code: code, HttpStatus: getHttpStatus(code), Params: params, Err: cause}
}

// Convenience constructors

func NewForbidden() error           { return NewError(i18n.CodeForbidden) }
func NewRateLimit() error           { return NewError(i18n.CodeRateLimit) }
func NewMaintenance() error         { return NewError(i18n.CodeMaintenance) }
func NewInternal(cause error) error { return NewErrorWithCause(i18n.CodeInternalError, cause) }
