package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (db *DB) OpenTransaction(ctx context.Context) (pgx.Tx, error) {
	tx, err := db.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("BeginTx CreateCounterMetric failed, %w", err)
	}
	return tx, nil
}

func (db *DB) Rollback(ctx context.Context, tx pgx.Tx) error {
	err := tx.Rollback(ctx)
	if err != nil {
		return fmt.Errorf("rollback failed, %w", err)
	}
	return nil
}

func (db *DB) Commit(ctx context.Context, tx pgx.Tx) error {
	err := tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit failed, %w", err)
	}
	return nil
}
