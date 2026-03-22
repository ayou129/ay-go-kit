package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ay/go-kit/ctxutil"
)

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

// Logger holds the configured loggers
type Logger struct {
	level    string
	debugLog *log.Logger
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
	logFile  *os.File
}

// Config for logger initialization
type Config struct {
	Level string // debug, info, warn, error
	Path  string // directory for log files
}

// ConfigFromEnv reads logger config from environment
func ConfigFromEnv() Config {
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if level == "" {
		level = LevelInfo
	}
	path := os.Getenv("LOG_PATH")
	if path == "" {
		path = "./logs"
	}
	return Config{Level: level, Path: path}
}

// New creates a new Logger
func New(cfg Config) (*Logger, error) {
	if err := os.MkdirAll(cfg.Path, 0755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	logFileName := filepath.Join(cfg.Path, fmt.Sprintf("app_%s.log", time.Now().Format("2006-01-02")))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	writer := io.MultiWriter(os.Stdout, logFile)
	return &Logger{
		level:    cfg.Level,
		debugLog: log.New(writer, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
		infoLog:  log.New(writer, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLog:  log.New(writer, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog: log.New(writer, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
		logFile:  logFile,
	}, nil
}

// Close closes the log file
func (l *Logger) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
}

func (l *Logger) shouldLog(level string) bool {
	levels := map[string]int{LevelDebug: 0, LevelInfo: 1, LevelWarn: 2, LevelError: 3}
	return levels[level] >= levels[l.level]
}

func (l *Logger) formatMsg(ctx context.Context, format string, v ...any) string {
	msg := fmt.Sprintf(format, v...)
	if ctx != nil {
		if traceID := ctxutil.GetTraceID(ctx); traceID != "" {
			return fmt.Sprintf("[TraceID: %s] %s", traceID, msg)
		}
	}
	return msg
}

// Debug logs a message at debug level
func (l *Logger) Debug(ctx context.Context, format string, v ...any) {
	if !l.shouldLog(LevelDebug) {
		return
	}
	if err := l.debugLog.Output(2, l.formatMsg(ctx, format, v...)); err != nil {
		fmt.Fprintf(os.Stderr, "logger debug output failed: %v\n", err)
	}
}

// Info logs a message at info level
func (l *Logger) Info(ctx context.Context, format string, v ...any) {
	if !l.shouldLog(LevelInfo) {
		return
	}
	if err := l.infoLog.Output(2, l.formatMsg(ctx, format, v...)); err != nil {
		fmt.Fprintf(os.Stderr, "logger info output failed: %v\n", err)
	}
}

// Warn logs a message at warn level
func (l *Logger) Warn(ctx context.Context, format string, v ...any) {
	if !l.shouldLog(LevelWarn) {
		return
	}
	if err := l.warnLog.Output(2, l.formatMsg(ctx, format, v...)); err != nil {
		fmt.Fprintf(os.Stderr, "logger warn output failed: %v\n", err)
	}
}

// Error logs a message at error level
func (l *Logger) Error(ctx context.Context, format string, v ...any) {
	if !l.shouldLog(LevelError) {
		return
	}
	if err := l.errorLog.Output(2, l.formatMsg(ctx, format, v...)); err != nil {
		fmt.Fprintf(os.Stderr, "logger error output failed: %v\n", err)
	}
}

// Global logger convenience (set once at startup)
var global *Logger

// SetGlobal sets the global logger instance
func SetGlobal(l *Logger) { global = l }

// G returns the global logger instance
func G() *Logger { return global }

// Debug logs at debug level using the global logger
func Debug(ctx context.Context, format string, v ...any) {
	if global != nil {
		global.Debug(ctx, format, v...)
	}
}

// Info logs at info level using the global logger
func Info(ctx context.Context, format string, v ...any) {
	if global != nil {
		global.Info(ctx, format, v...)
	}
}

// Warn logs at warn level using the global logger
func Warn(ctx context.Context, format string, v ...any) {
	if global != nil {
		global.Warn(ctx, format, v...)
	}
}

// Error logs at error level using the global logger
func Error(ctx context.Context, format string, v ...any) {
	if global != nil {
		global.Error(ctx, format, v...)
	}
}
