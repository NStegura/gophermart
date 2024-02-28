package repo

import (
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func (db *DB) runMigrations() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		return fmt.Errorf("failed to set db dialect, %w", err)
	}

	dbFromPool := stdlib.OpenDBFromPool(db.pool)
	if err := goose.Up(dbFromPool, "migrations"); err != nil {
		return fmt.Errorf("failed to migrate, %w", err)
	}
	return nil
}
