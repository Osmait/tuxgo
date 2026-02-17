package history

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func GetDBPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".local", "share", "tuxgo", "history.db")
}

func ensureDataDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dataDir := filepath.Join(homeDir, ".local", "share", "tuxgo")
	return os.MkdirAll(dataDir, 0755)
}

func openDB() (*sql.DB, error) {
	if err := ensureDataDir(); err != nil {
		return nil, err
	}

	dbPath := GetDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if err := createSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func createSchema(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			last_used DATETIME NOT NULL,
			use_count INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_history_name ON history(name)`,
		`CREATE INDEX IF NOT EXISTS idx_history_last_used ON history(last_used DESC)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}

	return nil
}
