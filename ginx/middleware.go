package ginx

import (
	"os"
	"strings"
	"time"

	"github.com/ay/go-kit/ctxutil"
	"github.com/ay/go-kit/token"
	"github.com/gin-gonic/gin"
)

// LogFunc is called after each request with method, path, status, durationMs
type LogFunc func(method, path string, status int, durationMs float64, traceID string)

// CommonMiddleware returns CORS + TraceID + Lang + request logging middleware
func CommonMiddleware(logFn LogFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/favicon.ico" {
			c.AbortWithStatus(204)
			return
		}

		// CORS
		env := os.Getenv("ENV")
		origin := c.Request.Header.Get("Origin")
		if strings.ToLower(env) != "prod" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			allowList := strings.Split(strings.TrimSpace(os.Getenv("CORS_ORIGINS")), ",")
			allowCreds := strings.EqualFold(os.Getenv("CORS_ALLOW_CREDENTIALS"), "true")

			matched := false
			if origin != "" {
				for _, o := range allowList {
					o = strings.TrimSpace(o)
					if o == "" {
						continue
					}
					if o == "*" || strings.EqualFold(o, origin) {
						matched = true
						break
					}
				}
			}
			if matched {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Add("Vary", "Origin")
				if allowCreds {
					c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			} else {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token, accept, origin, Cache-Control, X-Requested-With, Token-Access, Token-Refresh, X-Skip-Auth-Refresh")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Trace-Id, Server, Content-Disposition")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// TraceID
		traceID := token.GenerateTraceID()

		// Lang
		lang := c.Query("lang")
		if lang == "" {
			lang = ctxutil.DefaultLang
		}

		// Token headers
		accessToken := c.GetHeader("Token-Access")
		refreshToken := c.GetHeader("Token-Refresh")

		// Inject context
		ctx := c.Request.Context()
		ctx = ctxutil.WithTraceID(ctx, traceID)
		ctx = ctxutil.WithLang(ctx, lang)
		if accessToken != "" {
			ctx = ctxutil.WithAccessToken(ctx, accessToken)
		}
		if refreshToken != "" {
			ctx = ctxutil.WithRefreshToken(ctx, refreshToken)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Set("lang", lang)

		// Response headers
		c.Writer.Header().Set("Trace-Id", traceID)
		if author := os.Getenv("AUTHOR"); author != "" {
			c.Writer.Header().Set("Server", author)
		}

		// Process + log
		startTime := time.Now()
		c.Next()

		if logFn != nil {
			durationMs := float64(time.Since(startTime).Microseconds()) / 1000.0
			logFn(c.Request.Method, c.Request.URL.Path, c.Writer.Status(), durationMs, traceID)
		}
	}
}

// ErrorHandlerMiddleware catches errors set via c.Error() and writes them as API responses
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			WriteError(c, err)
		}
	}
}
