package store

import (
	"context"
	"database/sql"
	_ "embed"
	"golang-template-htmx-alpine/apps/todo/config"
	"golang-template-htmx-alpine/apps/todo/gen/db"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var ddl string

// Init creates a new in-memory SQLite database and runs the schema.sql file to create the tables
// It returns a new db.Queries instance connected to the in-memory database
func Init(config *config.Config) (*db.Queries, error) {
	ctx := context.Background()
	var sqlite *sql.DB
	var err error

	if config.Env == "test" {
		sqlite, err = sql.Open("sqlite3", ":memory:")
	} else {
		sqlite, err = sql.Open("sqlite3", "./todo.db")
	}

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
