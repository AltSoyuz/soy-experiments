package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestApplyMigrations(t *testing.T) {
	// Create an in-memory SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory SQLite database: %v", err)
	}
	defer db.Close()

	// Create migration applier
	migrationApp := &migrationApplier{Db: db}

	// Apply all migrations
	err = migrationApp.applyMigrations()
	if err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	// Verify that the migrations were applied successfully
	// For example, check if a table created by the migrations exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='user'").Scan(&tableName)
	if err != nil {
		t.Fatalf("failed to verify migration: %v", err)
	}

	if tableName != "user" {
		t.Fatalf("expected table 'your_table_name' to exist, but it does not")
	}
}
