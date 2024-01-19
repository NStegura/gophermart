package repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/NStegura/gophermart/internal/customerrors"
	"github.com/NStegura/gophermart/internal/repo/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type DB struct {
	pool *pgxpool.Pool

	logger *logrus.Logger
}

func New(ctx context.Context, dsn string, logger *logrus.Logger) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}
	return &DB{
		pool:   pool,
		logger: logger,
	}, nil
}

func (db *DB) Shutdown(_ context.Context) {
	db.logger.Debug("db shutdown")
	db.pool.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	db.logger.Debug("Ping db")
	err := db.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("DB ping eror, %w", err)
	}
	return nil
}

func (db *DB) GetUser(ctx context.Context, tx pgx.Tx, login string) (u models.User, err error) {
	const query = `
		SELECT u.id, u.login, u.password, u.salt, u.created_at
		FROM "user" u
		WHERE u.login = $1; 
	`
	err = tx.QueryRow(ctx, query, login).Scan(
		&u.ID,
		&u.Login,
		&u.Password,
		&u.Salt,
		&u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = customerrors.ErrNotFound
			return
		}
		return u, fmt.Errorf("get user failed, %w", err)
	}

	return u, nil
}

func (db *DB) CreateUser(ctx context.Context, tx pgx.Tx, login, password, salt string) (ID int64, err error) {
	const query = `
		INSERT INTO "user" (login, password, salt)
		VALUES ($1, $2, $3)
		RETURNING  "user".id; 
	`

	err = tx.QueryRow(ctx, query,
		login, password, salt,
	).Scan(&ID)

	if err != nil {
		return ID, fmt.Errorf("CreateUser failed, %w", err)
	}
	db.logger.Debugf("Create user, id, %v", ID)
	return
}
