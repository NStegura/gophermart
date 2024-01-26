package accrualsync

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/NStegura/gophermart/internal/repo/models"
)

type Repository interface {
	Ping(ctx context.Context) error

	GetUserByID(ctx context.Context, tx pgx.Tx, ID int64, forUpdate bool) (u models.User, err error)
	UpdateUserBalance(ctx context.Context, tx pgx.Tx, userID int64, balance, withdrawn float64) (err error)
	GetOrder(ctx context.Context, tx pgx.Tx, orderID int64, forUpdate bool) (o models.Order, err error)
	UpdateOrder(ctx context.Context, tx pgx.Tx, orderID int64, accrual float64, status string) error
	GetNotProcessedOrders(ctx context.Context, tx pgx.Tx) ([]models.Order, error)

	OpenTransaction(ctx context.Context) (tx pgx.Tx, err error)
	Rollback(ctx context.Context, tx pgx.Tx) error
	Commit(ctx context.Context, tx pgx.Tx) error
}
