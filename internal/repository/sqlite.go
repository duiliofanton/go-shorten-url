package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/duiliofanton/go-shorten-url/internal/config"
	"github.com/duiliofanton/go-shorten-url/internal/models"
	"github.com/mattn/go-sqlite3"
)

var ErrShortCodeConflict = errors.New("short code already exists")

type SQLiteURLRepository struct {
	db *sql.DB
}

func NewSQLiteURLRepository(cfg config.DatabaseConfig) (*SQLiteURLRepository, error) {
	if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	connStr := fmt.Sprintf("%s/%s?cache=shared&mode=rwc&_loc=auto", cfg.Dir, cfg.Name)

	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	return &SQLiteURLRepository{db: db}, nil
}

func (r *SQLiteURLRepository) Create(ctx context.Context, url *models.URL) error {
	query := `
		INSERT INTO urls (id, original, short_code, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, url.ID, url.Original, url.ShortCode, url.CreatedAt, url.UpdatedAt)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return ErrShortCodeConflict
		}
		return fmt.Errorf("failed to create url: %w", err)
	}
	return nil
}

func (r *SQLiteURLRepository) GetByID(ctx context.Context, id string) (*models.URL, error) {
	query := `
		SELECT id, original, short_code, created_at, updated_at
		FROM urls
		WHERE id = ?
	`
	var url models.URL
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&url.ID, &url.Original, &url.ShortCode, &url.CreatedAt, &url.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (r *SQLiteURLRepository) GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	query := `
		SELECT id, original, short_code, created_at, updated_at
		FROM urls
		WHERE short_code = ?
	`
	var url models.URL
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(
		&url.ID, &url.Original, &url.ShortCode, &url.CreatedAt, &url.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (r *SQLiteURLRepository) Update(ctx context.Context, url *models.URL) error {
	query := `
		UPDATE urls
		SET original = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, url.Original, url.UpdatedAt, url.ID)
	if err != nil {
		return fmt.Errorf("failed to update url: %w", err)
	}
	return nil
}

func (r *SQLiteURLRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM urls WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete url: %w", err)
	}
	return nil
}

func (r *SQLiteURLRepository) List(ctx context.Context, page, perPage int) ([]models.URL, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM urls`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count urls: %w", err)
	}

	offset := (page - 1) * perPage
	query := `
		SELECT id, original, short_code, created_at, updated_at
		FROM urls
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list urls: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("failed to close rows", "error", err)
		}
	}()

	urls := []models.URL{}
	for rows.Next() {
		var url models.URL
		if err := rows.Scan(&url.ID, &url.Original, &url.ShortCode, &url.CreatedAt, &url.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan url: %w", err)
		}
		urls = append(urls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate urls: %w", err)
	}

	return urls, total, nil
}

func (r *SQLiteURLRepository) Close() {
	if err := r.db.Close(); err != nil {
		slog.Error("failed to close database", "error", err)
	}
}

func (r *SQLiteURLRepository) DB() *sql.DB {
	return r.db
}
