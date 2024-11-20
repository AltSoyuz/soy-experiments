package store

import (
	"context"
	"database/sql"
	_ "embed"
	"golang-template-htmx-alpine/apps/todo/gen/db"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var ddl string

// Init creates a new in-memory SQLite database and runs the schema.sql file to create the tables
// It returns a new db.Queries instance connected to the in-memory database
func Init() (*db.Queries, error) {
	ctx := context.Background()
	sqlite, err := sql.Open("sqlite3", ":memory:")

	if err != nil {
		return nil, err
	}

	_, err = sqlite.ExecContext(ctx, ddl)
	if err != nil {
		return nil, err
	}

	q := db.New(sqlite)

	return q, nil
}
