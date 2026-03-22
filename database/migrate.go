package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// MigrateLogFunc for migration logging (optional)
type MigrateLogFunc func(format string, args ...any)

// RunMigrations executes all .sql files in migrationsDir (ordered by filename).
// Uses schema_migrations table to track executed files.
func RunMigrations(db *gorm.DB, migrationsDir string, logFn MigrateLogFunc) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`).Error; err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var sqlFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			sqlFiles = append(sqlFiles, e.Name())
		}
	}
	sort.Strings(sqlFiles)

	for _, fileName := range sqlFiles {
		var count int64
		db.Raw("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", fileName).Scan(&count)
		if count > 0 {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, fileName))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", fileName, err)
		}

		if logFn != nil {
			logFn("executing migration: %s", fileName)
		}

		if err := db.Exec(string(content)).Error; err != nil {
			return fmt.Errorf("execute migration %s: %w", fileName, err)
		}

		if err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", fileName).Error; err != nil {
			return fmt.Errorf("record migration %s: %w", fileName, err)
		}

		if logFn != nil {
			logFn("migration %s completed", fileName)
		}
	}

	return nil
}
