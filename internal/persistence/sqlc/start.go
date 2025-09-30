package sqlc

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var schemaLocation = "../sql/schema.sql"
var schema embed.FS

func Start(ctx context.Context, path string) (*Queries, *sql.DB, error) {
	sqlDB, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxLifetime(time.Minute)

	schema, _ := schema.ReadFile(schemaLocation)

	if _, err := sqlDB.ExecContext(ctx, string(schema)); err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to create schema: %w", err)
	}

	queries, err := Prepare(ctx, sqlDB)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare queries: %w", err)
	}

	return queries, sqlDB, nil
}
