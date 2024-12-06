package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AltSoyuz/soy-experiments/apps/todo/config"
	"github.com/AltSoyuz/soy-experiments/apps/todo/gen/db"

	_ "github.com/mattn/go-sqlite3"
)

type migrationApplier struct {
	Db *sql.DB
}

func (ma *migrationApplier) applyMigrations() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	migrationsDir := filepath.Join(cwd, "../store/migrations")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrationPaths []string
	for _, entry := range entries {
		if !entry.IsDir() &&
			strings.HasSuffix(entry.Name(), ".sql") &&
			!strings.HasSuffix(entry.Name(), ".down.sql") {
			migrationPaths = append(migrationPaths, filepath.Join(migrationsDir, entry.Name()))
		}
	}

	sort.Strings(migrationPaths)

	for _, path := range migrationPaths {
		migrationBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		_, err = ma.Db.Exec(string(migrationBytes))
		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", path, err)
		}
	}

	return nil
}

// Init creates a new in-memory SQLite database and runs the schema.sql file to create the tables
// It returns a new db.Queries instance connected to the in-memory database
func Init(config *config.Config) (*db.Queries, error) {
	var sqlite *sql.DB
	var err error

	if config.Env == "test" {
		sqlite, err = sql.Open("sqlite3", ":memory:")
		if err != nil {
			return nil, err
		}

		// Create migration applier and apply migrations only in test
		migrationApp := &migrationApplier{Db: sqlite}
		if err := migrationApp.applyMigrations(); err != nil {
			return nil, fmt.Errorf("migration failed: %w", err)
		}
	} else {
		sqlite, err = sql.Open("sqlite3", "./todo.db")
		if err != nil {
			return nil, err
		}
	}

	q := db.New(sqlite)
	err = q.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return q, nil
}
