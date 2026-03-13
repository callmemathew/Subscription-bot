package storage

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDB(path string) (*sql.DB, error) {
	if path == "" {
		return nil, errors.New("empty db path")
	}

	cleanPath := filepath.Clean(path)

	if _, err := os.Stat(cleanPath); err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("database file not found: " + cleanPath)
		}
		return nil, err
	}

	db, err := sql.Open("sqlite3", cleanPath)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS clients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		purchase_date TEXT NOT NULL,
		expire_date TEXT,
		notified_7 INTEGER NOT NULL DEFAULT 0
	);`

	_, err = db.Exec(schema)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
