package database

import (
	"fmt"
	"os"
	"strconv"
	"time"

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
