package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Open(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) Migrate() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS videos (
			id          TEXT PRIMARY KEY,
			title       TEXT NOT NULL,
			file_path   TEXT NOT NULL UNIQUE,
			file_size   INTEGER NOT NULL DEFAULT 0,
			duration    REAL NOT NULL DEFAULT 0,
			width       INTEGER NOT NULL DEFAULT 0,
			height      INTEGER NOT NULL DEFAULT 0,
			codec       TEXT NOT NULL DEFAULT '',
			bitrate     INTEGER NOT NULL DEFAULT 0,
			container   TEXT NOT NULL DEFAULT '',
			created_at  TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
		);

		CREATE INDEX IF NOT EXISTS idx_videos_title ON videos(title);
		CREATE INDEX IF NOT EXISTS idx_videos_path  ON videos(file_path);
	`)
	return err
}
