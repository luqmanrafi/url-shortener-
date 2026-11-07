package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/luqmanrafi/urlshortener/internal/model"
	_ "modernc.org/sqlite"
)

type SQLite struct {
	db *sql.DB
}

func NewSQLite(path string)(*SQLite, error){
	db, err := sql.Open("sqlite", path)
	if err != nil{
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT NOT NULL UNIQUE,
		original_url TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		click_count INTEGER NOT NULL DEFAULT 0,
		expires_at DATETIME  
	);`

	if _, err := db.Exec(schema); err != nil{
		return nil, err
	}
	// Migrate legacy column name typo if present
	_, _ = db.Exec(`ALTER TABLE urls RENAME COLUMN clik_count TO click_count;`)
	_, _ = db.Exec(`ALTER TABLE expires_at ADD COLUMN expire_at DATETIME`)
	return &SQLite{db: db}, nil
}


func (r *SQLite) Save(ctx context.Context, code, originalURL string, expiresAt *time.Time)error{
	_, err := r.db.ExecContext(ctx, 
	`INSERT INTO urls (code, original_url, expires_at) VALUES (?, ?, ?)`, code, originalURL, expiresAt)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return model.ErrCodeExists
		}
		return err
	}
	return nil
}

func (r *SQLite) FindByCode(ctx context.Context, code string)(*model.URL, error){
	var u model.URL
	var expires sql.NullTime
	err := r.db.QueryRowContext(ctx, 
	`SELECT code, original_url, expires_at FROM urls WHERE code = ?`, code).Scan(&u.Code, &u.Original, &u.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows){
		return nil, model.ErrNotFound
	}
	if err != nil{
		return nil, err
	}
	if expires.Valid{
		u.ExpiresAt = &expires.Time
	}
	return &u, nil
}

func (r *SQLite) IncrementClick(ctx context.Context, code string) error {
	res, err := r.db.ExecContext(ctx, 
		`UPDATE urls SET click_count = click_count + 1 WHERE code = ?`, code)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *SQLite) GetStats(ctx context.Context, code string) (*model.URL, error) {
	var item model.URL
	err := r.db.QueryRowContext(ctx, 
		`SELECT code, original_url, click_count, created_at FROM urls WHERE code = ?`, code,
	).Scan(&item.Code, &item.Original, &item.ClickCount, &item.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SQLite) Ping(ctx context.Context) error{
	return r.db.PingContext(ctx)
}

