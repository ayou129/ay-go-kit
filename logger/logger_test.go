package logger

import (
	"context"
	"strings"
	"testing"

	"github.com/ay/go-kit/ctxutil"
)

func TestShouldLog_LevelOrdering(t *testing.T) {
	tests := []struct {
		loggerLevel string
		msgLevel    string
		want        bool
	}{
		{"debug", "debug", true},
		{"debug", "info", true},
		{"debug", "warn", true},
		{"debug", "error", true},
		{"info", "debug", false},
		{"info", "info", true},
		{"info", "warn", true},
		{"info", "error", true},
		{"warn", "debug", false},
		{"warn", "info", false},
		{"warn", "warn", true},
		{"warn", "error", true},
		{"error", "debug", false},
		{"error", "info", false},
		{"error", "warn", false},
		{"error", "error", true},
	}

	for _, tt := range tests {
		l := &Logger{level: tt.loggerLevel}
		got := l.shouldLog(tt.msgLevel)
		if got != tt.want {
			t.Errorf("logger=%s msg=%s: shouldLog=%v, want %v", tt.loggerLevel, tt.msgLevel, got, tt.want)
		}
	}
}

func TestFormatMsg_WithTraceID(t *testing.T) {
	l := &Logger{level: LevelDebug}
	ctx := ctxutil.WithTraceID(context.Background(), "abc-123")

	msg := l.formatMsg(ctx, "hello %s", "world")
	if !strings.Contains(msg, "[TraceID: abc-123]") {
		t.Errorf("expected TraceID in message, got: %s", msg)
	}
	if !strings.Contains(msg, "hello world") {
		t.Errorf("expected formatted message, got: %s", msg)
	}
}

func TestFormatMsg_WithoutTraceID(t *testing.T) {
	l := &Logger{level: LevelDebug}
	ctx := context.Background()

	msg := l.formatMsg(ctx, "plain %d", 42)
	if strings.Contains(msg, "TraceID") {
		t.Errorf("unexpected TraceID in message: %s", msg)
	}
	if msg != "plain 42" {
		t.Errorf("got %q, want %q", msg, "plain 42")
	}
}

func TestFormatMsg_NilContext(t *testing.T) {
	l := &Logger{level: LevelDebug}

	msg := l.formatMsg(nil, "no ctx")
	if msg != "no ctx" {
		t.Errorf("got %q, want %q", msg, "no ctx")
	}
}

func TestNew_CreatesLoggerSuccessfully(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{Level: LevelInfo, Path: tmpDir}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer l.Close()

	if l.level != LevelInfo {
		t.Errorf("level = %s, want %s", l.level, LevelInfo)
	}
	if l.debugLog == nil || l.infoLog == nil || l.warnLog == nil || l.errorLog == nil {
		t.Error("one or more internal loggers are nil")
	}
	if l.logFile == nil {
		t.Error("logFile is nil")
	}
}

func TestGlobalFunctions_NilGlobal_NoPanic(t *testing.T) {
	// Ensure global is nil
	old := global
	global = nil
	defer func() { global = old }()

	// These should not panic
	ctx := context.Background()
	Debug(ctx, "test %d", 1)
	Info(ctx, "test %d", 2)
	Warn(ctx, "test %d", 3)
	Error(ctx, "test %d", 4)
}
