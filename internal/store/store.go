package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func NewDB(ctx context.Context, dbDriver string, dbUrl string, migrate bool) (*Queries, error) {
	sqldb, err := sql.Open(dbDriver, dbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s database: %w", dbDriver, err)
	}

	if migrate {
		goose.SetBaseFS(migrations)

		if err := goose.SetDialect(dbDriver); err != nil {
			return nil, fmt.Errorf("failed to set goose dialect: %w", err)
		}

		if err := goose.Up(sqldb, "migrations"); err != nil {
			return nil, fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	return New(sqldb), nil
}
