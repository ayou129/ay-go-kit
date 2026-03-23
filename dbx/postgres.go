package dbx

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ay/go-kit/ctxutil"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config holds PostgreSQL connection configuration
type Config struct {
	DSN             string
	MaxIdle         int
	MaxOpen         int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// ConfigFromEnv reads PostgreSQL config from environment variables
func ConfigFromEnv() Config {
	cfg := Config{
		DSN:             os.Getenv("POSTGRES_DSN"),
		MaxIdle:         10,
		MaxOpen:         100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	if cfg.DSN == "" {
		sslMode := os.Getenv("POSTGRES_SSLMODE")
		if sslMode == "" {
			sslMode = "disable"
		}
		timeZone := os.Getenv("TIMEZONE")
		if timeZone == "" {
			timeZone = "Asia/Shanghai"
		}
		cfg.DSN = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"),
			os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_DATABASE"), sslMode, timeZone,
		)
	}

	if v, _ := strconv.Atoi(os.Getenv("POSTGRES_MAX_IDLE")); v > 0 {
		cfg.MaxIdle = v
	}
	if v, _ := strconv.Atoi(os.Getenv("POSTGRES_MAX_OPEN")); v > 0 {
		cfg.MaxOpen = v
	}

	return cfg
}

// SQLLogFunc is called after each SQL execution for logging
type SQLLogFunc func(sql string, rows int64, traceID string, err error)

// Open creates a new GORM database connection
func Open(cfg Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return db, nil
}

// RegisterSQLCallbacks registers GORM callbacks for SQL logging
func RegisterSQLCallbacks(db *gorm.DB, logFn SQLLogFunc) error {
	callback := func(gdb *gorm.DB) {
		sql := gdb.Dialector.Explain(gdb.Statement.SQL.String(), gdb.Statement.Vars...)
		rows := gdb.Statement.RowsAffected
		traceID := ""
		if ctx := gdb.Statement.Context; ctx != nil {
			traceID = ctxutil.GetTraceID(ctx)
		}
		logFn(sql, rows, traceID, gdb.Error)
	}

	type registrar struct {
		register func(string, func(*gorm.DB)) error
		name     string
	}
	for _, r := range []registrar{
		{db.Callback().Query().After("gorm:query").Register, "query"},
		{db.Callback().Create().After("gorm:create").Register, "create"},
		{db.Callback().Update().After("gorm:update").Register, "update"},
		{db.Callback().Delete().After("gorm:delete").Register, "delete"},
		{db.Callback().Row().After("gorm:row").Register, "row"},
		{db.Callback().Raw().After("gorm:raw").Register, "raw"},
	} {
		if err := r.register("gokit:log_sql", callback); err != nil {
			return fmt.Errorf("register %s callback: %w", r.name, err)
		}
	}
	return nil
}


// Close closes the database connection
func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
