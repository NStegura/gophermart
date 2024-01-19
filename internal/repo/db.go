package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/NStegura/gophermart/internal/customerrors"
	"github.com/NStegura/gophermart/internal/repo/models"
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

func (db *DB) CreateUser(ctx context.Context, tx pgx.Tx, login, password, salt string) (id int64, err error) {
	const query = `
		INSERT INTO "user" (login, password, salt)
		VALUES ($1, $2, $3)
		RETURNING  "user".id; 
	`

	err = tx.QueryRow(ctx, query,
		login, password, salt,
	).Scan(&id)

	if err != nil {
		return id, fmt.Errorf("CreateUser failed, %w", err)
	}
	db.logger.Debugf("Create user, id, %v", id)
	return
}

func (db *DB) GetOrder(ctx context.Context, tx pgx.Tx, orderID int64) (o models.Order, err error) {
	const query = `
		SELECT o.id, o.status, o.user_id, o.created_at
		FROM "order" o
		WHERE o.id = $1; 
	`
	err = tx.QueryRow(ctx, query, orderID).Scan(
		&o.ID,
		&o.Status,
		&o.UserID,
		&o.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = customerrors.ErrNotFound
			return
		}
		return o, fmt.Errorf("get user failed, %w", err)
	}

	return o, nil
}

func (db *DB) CreateOrder(ctx context.Context, tx pgx.Tx, userID, orderID int64) (err error) {
	var id int64

	const query = `
		INSERT INTO "order" (id, status, user_id)
		VALUES ($1, $2, $3)
		RETURNING  "order".id; 
	`

	err = tx.QueryRow(ctx, query,
		orderID, "NEW", userID,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("CreateUser failed, %w", err)
	}
	db.logger.Debugf("Create order, id, %v", id)
	return
}
