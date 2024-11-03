package queries

import (
	"context"
	"database/sql"
	_ "embed"
	_ "github.com/mattn/go-sqlite3"
	"golang-template-htmx-alpine/gen/db"
)

//go:embed schema.sql
var ddl string

func New() (*db.Queries, error) {
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
