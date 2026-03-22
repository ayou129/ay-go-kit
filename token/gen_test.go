package token

import (
	"strings"
	"testing"
)

func TestGenerateToken_Length(t *testing.T) {
	tok := GenerateToken()
	if len(tok) != 48 {
		t.Errorf("GenerateToken() length = %d, want 48", len(tok))
	}
}

func TestGenerateTraceID_Length(t *testing.T) {
	tid := GenerateTraceID()
	if len(tid) != 32 {
		t.Errorf("GenerateTraceID() length = %d, want 32", len(tid))
	}
}

func TestGenerate_CustomLengths(t *testing.T) {
	tests := []struct {
		name      string
		timeLen   int
		randomLen int
		wantLen   int
	}{
		{"5+10", 5, 10, 15},
		{"0+20", 0, 20, 20},
		{"20+0", 20, 0, 20},
		{"1+1", 1, 1, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Generate(tt.timeLen, tt.randomLen)
			if len(got) != tt.wantLen {
				t.Errorf("Generate(%d, %d) length = %d, want %d", tt.timeLen, tt.randomLen, len(got), tt.wantLen)
			}
		})
	}
}

func TestGenerate_ZeroZero(t *testing.T) {
	got := Generate(0, 0)
	if got != "" {
		t.Errorf("Generate(0, 0) = %q, want empty", got)
	}
}

func TestGenerateToken_Uniqueness(t *testing.T) {
	a := GenerateToken()
	b := GenerateToken()
	if a == b {
		t.Errorf("two GenerateToken() calls returned the same value: %q", a)
	}
}

func TestGenerateTraceID_Uniqueness(t *testing.T) {
	a := GenerateTraceID()
	b := GenerateTraceID()
	if a == b {
		t.Errorf("two GenerateTraceID() calls returned the same value: %q", a)
	}
}

func TestGenerateToken_Charset(t *testing.T) {
	tok := GenerateToken()
	// time prefix is digits, random suffix is from charset (alphanumeric)
	for i, c := range tok {
		if !strings.ContainsRune(charset+"0123456789", c) {
			t.Errorf("GenerateToken()[%d] = %q, not in allowed charset", i, string(c))
		}
	}
}

func TestRandomString_Charset(t *testing.T) {
	s := randomString(100)
	for i, c := range s {
		if !strings.ContainsRune(charset, c) {
			t.Errorf("randomString[%d] = %q, not in charset", i, string(c))
		}
	}
}

func TestRandomString_NegativeLength(t *testing.T) {
	if got := randomString(-1); got != "" {
		t.Errorf("randomString(-1) = %q, want empty", got)
	}
}

func TestTimeString_NegativeLength(t *testing.T) {
	if got := timeString(-1); got != "" {
		t.Errorf("timeString(-1) = %q, want empty", got)
	}
}
