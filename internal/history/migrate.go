package history

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type yamlHistory struct {
	Entries []yamlEntry `yaml:"entries"`
}

type yamlEntry struct {
	Path     string    `yaml:"path"`
	Name     string    `yaml:"name"`
	LastUsed time.Time `yaml:"last_used"`
	UseCount int       `yaml:"use_count"`
}

func migrateFromYAML(db *sql.DB) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	yamlPath := filepath.Join(homeDir, ".local", "share", "tuxgo", "history.yaml")

	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		return nil
	}

	migratedPath := yamlPath + ".migrated"
	if _, err := os.Stat(migratedPath); err == nil {
		return nil
	}

	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil
	}

	var yh yamlHistory
	if err := yaml.Unmarshal(data, &yh); err != nil {
		return nil
	}

	if len(yh.Entries) == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO history (path, name, last_used, use_count)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range yh.Entries {
		if e.Path == "" || e.Name == "" {
			continue
		}
		_, err := stmt.Exec(e.Path, e.Name, e.LastUsed, e.UseCount)
		if err != nil {
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return os.Rename(yamlPath, migratedPath)
}
