package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (db *DB) GetUserByLogin(ctx context.Context, tx pgx.Tx, login string) (u models.User, err error) {
	const query = `
		SELECT u.id, u.login, u.password, u.salt, u.balance, u.withdrawn, u.created_at
		FROM "user" u
		WHERE u.login = $1; 
	`
	err = tx.QueryRow(ctx, query, login).Scan(
		&u.ID,
		&u.Login,
		&u.Password,
		&u.Salt,
		&u.Balance,
		&u.Withdrawn,
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

func (db *DB) GetUserByID(ctx context.Context, tx pgx.Tx, ID int64, forUpdate bool) (u models.User, err error) {
	var query string
	if forUpdate {
		query = `
		SELECT u.id, u.login, u.password, u.salt, u.balance, u.withdrawn, u.created_at
		FROM "user" u
		WHERE u.id = $1
		FOR UPDATE; 
	`
	} else {
		query = `
		SELECT u.id, u.login, u.password, u.salt, u.balance, u.withdrawn, u.created_at
		FROM "user" u
		WHERE u.id = $1; 
	`
	}

	err = tx.QueryRow(ctx, query, ID).Scan(
		&u.ID,
		&u.Login,
		&u.Password,
		&u.Salt,
		&u.Balance,
		&u.Withdrawn,
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

func (db *DB) UpdateUserBalance(ctx context.Context, tx pgx.Tx, balance, withdrawn float64) (err error) {
	var id int64
	const query = `
		UPDATE "user"
		SET balance = $1, withdrawn = $2, updated_at = $3
		RETURNING  "user".id; 
	`

	err = tx.QueryRow(ctx, query,
		balance,
		withdrawn,
		time.Now(),
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("UpdateUserBalance failed, %w", err)
	}
	db.logger.Debugf("Update user balance, id, %v", id)
	return
}

func (db *DB) GetOrder(ctx context.Context, tx pgx.Tx, orderID int64) (o models.Order, err error) {
	const query = `
		SELECT o.id, o.status, o.user_id, o.created_at, o.updated_at
		FROM "order" o
		WHERE o.id = $1; 
	`
	err = tx.QueryRow(ctx, query, orderID).Scan(
		&o.ID,
		&o.Status,
		&o.UserID,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = customerrors.ErrNotFound
			return
		}
		return o, fmt.Errorf("get order failed, %w", err)
	}

	return o, nil
}

func (db *DB) GetOrders(ctx context.Context, tx pgx.Tx, userID int64) (orders []models.Order, err error) {
	var rows pgx.Rows

	const query = `
		SELECT o.id, o.status, o.user_id, o.accrual, o.created_at, o.updated_at
		FROM "order" o
		WHERE o.user_id = $1
		ORDER BY o.created_at; 
	`
	rows, err = tx.Query(ctx, query, userID)
	if err != nil {
		return orders, fmt.Errorf("get orders failed, %w", err)
	}

	for rows.Next() {
		var o models.Order
		err = rows.Scan(
			&o.ID,
			&o.Status,
			&o.UserID,
			&o.Accrual,
			&o.CreatedAt,
			&o.UpdatedAt,
		)
		if err != nil {
			db.logger.Debug(err)
			return orders, fmt.Errorf("get orders failed, %w", err)
		}
		db.logger.Debug(o)
		orders = append(orders, o)
	}
	if err = rows.Err(); err != nil {
		return orders, fmt.Errorf("get orders failed, %w", err)
	}

	return orders, nil
}

func (db *DB) CreateOrder(ctx context.Context, tx pgx.Tx, userID, orderID int64) (err error) {
	var id int64

	const query = `
		INSERT INTO "order" (id, status, user_id)
		VALUES ($1, $2, $3)
		RETURNING  "order".id; 
	`

	err = tx.QueryRow(ctx, query,
		orderID, NEW.String(), userID,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("CreateOrder failed, %w", err)
	}
	db.logger.Debugf("Create order, id, %v", id)
	return
}

func (db *DB) CreateWithdraw(ctx context.Context, tx pgx.Tx, userID, orderID int64, sum float64) (err error) {
	var id int64

	const query = `
		INSERT INTO "withdraw" (order_id, user_id, sum)
		VALUES ($1, $2, $3)
		RETURNING  "withdraw".id; 
	`

	err = tx.QueryRow(ctx, query,
		orderID, userID, sum,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("CreateWithdraw failed, %w", err)
	}
	db.logger.Debugf("Create withdraw, id, %v", id)
	return
}

func (db *DB) GetWithdrawals(ctx context.Context, tx pgx.Tx, userID int64) (withdrawals []models.Withdraw, err error) {
	var rows pgx.Rows

	const query = `
		SELECT w.id, w.order_id, w.user_id, w.sum, w.created_at
		FROM "withdraw" w
		WHERE w.user_id = $1
		ORDER BY w.created_at; 
	`
	rows, err = tx.Query(ctx, query, userID)
	if err != nil {
		return withdrawals, fmt.Errorf("get orders failed, %w", err)
	}

	for rows.Next() {
		var w models.Withdraw
		err = rows.Scan(
			&w.ID,
			&w.OrderID,
			&w.UserID,
			&w.Sum,
			&w.CreatedAt,
		)
		if err != nil {
			db.logger.Debug(err)
			return withdrawals, fmt.Errorf("get withdrawals failed, %w", err)
		}
		db.logger.Debug(w)
		withdrawals = append(withdrawals, w)
	}
	if err = rows.Err(); err != nil {
		return withdrawals, fmt.Errorf("get withdrawals failed, %w", err)
	}

	return withdrawals, nil
}
