package ginx

import (
	"net/http"

	"github.com/ay/go-kit/i18n"
	"github.com/gin-gonic/gin"
)

func getLang(c *gin.Context) string {
	if lang, exists := c.Get("lang"); exists {
		if s, ok := lang.(string); ok {
			return s
		}
	}
	return i18n.GetDefaultLang()
}

// Success writes a success response (data optional, defaults to empty object)
func Success(c *gin.Context, data ...any) {
	lang := getLang(c)
	var d any
	if len(data) > 0 {
		d = data[0]
	} else {
		d = struct{}{}
	}
	c.JSON(http.StatusOK, ApiResponse{
		Code: i18n.CodeSuccess,
		Msg:  i18n.GetLangMsg(i18n.CodeSuccess, lang),
		Data: d,
	})
}

// Error writes an error response with HTTP status and error code
func Error(c *gin.Context, httpStatus int, code int, params ...any) {
	lang := getLang(c)
	c.JSON(httpStatus, ApiResponse{
		Code: code,
		Msg:  i18n.GetLangMsg(code, lang, params...),
		Data: nil,
	})
}

// WriteError writes error response, handling BizError or generic error
func WriteError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if bizErr, ok := err.(BizError); ok {
		bizErr.WriteResponse(c)
		return
	}
	lang := getLang(c)
	c.JSON(http.StatusInternalServerError, ApiResponse{
		Code: i18n.CodeInternalError,
		Msg:  i18n.GetLangMsg(i18n.CodeInternalError, lang),
		Data: nil,
	})
	c.Abort()
}
