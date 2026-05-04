package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDB(path string) (*sql.DB, error) {
	if path == "" {
		return nil, errors.New("OpenDB: empty db path")
	}

	cleanPath := filepath.Clean(path)

	db, err := sql.Open("sqlite3", cleanPath)
	if err != nil {
		return nil, fmt.Errorf("OpenDB: sql.Open failed for path=%q: %w", cleanPath, err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS clients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		purchase_date TEXT NOT NULL,
		expire_date TEXT,
		notified_7 INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS payments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client_name TEXT NOT NULL,
		type TEXT NOT NULL,
		amount INTEGER NOT NULL,
		paid_at TEXT NOT NULL
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("OpenDB: schema exec failed: %w", err)
	}

	return db, nil
}
