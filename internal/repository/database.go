package repository

import (
	"context"
	"database/sql"
	"fmt"
)

func InitDatabase(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS urls (
		id TEXT PRIMARY KEY,
		original TEXT NOT NULL,
		short_code TEXT UNIQUE NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
	CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls(created_at DESC);
	`

	if _, err := db.ExecContext(context.Background(), schema); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	return nil
}
