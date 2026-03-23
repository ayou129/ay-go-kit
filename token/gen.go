package token

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(length int) string {
	if length <= 0 {
		return ""
	}
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
			continue
		}
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

func timeString(length int) string {
	if length <= 0 {
		return ""
	}
	ts := fmt.Sprintf("%d", time.Now().UnixNano())
	if len(ts) >= length {
		return ts[:length]
	}
	return fmt.Sprintf("%0*s", length, ts)
}

// Generate creates a token with time prefix + random suffix
func Generate(timeLen, randomLen int) string {
	return timeString(timeLen) + randomString(randomLen)
}

// GenerateToken creates a 48-char token (12 time + 36 random)
func GenerateToken() string { return Generate(12, 36) }

// GenerateTraceID creates a 32-char trace ID (8 time + 24 random)
func GenerateTraceID() string { return Generate(8, 24) }

// RandomString creates a cryptographically random string of given length
func RandomString(n int) string { return randomString(n) }
