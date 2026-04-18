package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gorm.io/gorm"
)

func ApplyMigrations(db *gorm.DB, dir string) error {
	if err := db.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
  id BIGSERIAL PRIMARY KEY,
  filename TEXT NOT NULL UNIQUE,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`).Error; err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		filename := filepath.Base(file)
		var count int64
		if err := db.Table("schema_migrations").Where("filename = ?", filename).Count(&count).Error; err != nil {
			return fmt.Errorf("check migration %s: %w", filename, err)
		}
		if count > 0 {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filename, err)
		}

		tx := db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("begin migration tx %s: %w", filename, tx.Error)
		}

		if err := tx.Exec(string(content)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", filename, err)
		}

		if err := tx.Exec("INSERT INTO schema_migrations (filename, applied_at) VALUES (?, ?)", filename, time.Now()).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", filename, err)
		}

		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("commit migration %s: %w", filename, err)
		}
	}
	return nil
}
