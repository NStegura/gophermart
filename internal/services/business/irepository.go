package business

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/NStegura/gophermart/internal/repo/models"
)

type Repository interface {
	Ping(ctx context.Context) error

	CreateUser(ctx context.Context, tx pgx.Tx, login, password, salt string) (id int64, err error)
	GetUserByLogin(ctx context.Context, tx pgx.Tx, login string) (u models.User, err error)
	GetUserByID(ctx context.Context, tx pgx.Tx, ID int64, forUpdate bool) (u models.User, err error)
	UpdateUserBalance(ctx context.Context, tx pgx.Tx, balance, withdrawn float64) (err error)
	GetOrder(ctx context.Context, tx pgx.Tx, orderID int64) (o models.Order, err error)
	GetOrders(ctx context.Context, tx pgx.Tx, userID int64) (orders []models.Order, err error)
	CreateOrder(ctx context.Context, tx pgx.Tx, userID, orderID int64) (err error)
	CreateWithdraw(ctx context.Context, tx pgx.Tx, userID, orderID int64, sum float64) (err error)
	GetWithdrawals(ctx context.Context, tx pgx.Tx, userID int64) (withdrawals []models.Withdraw, err error)

	OpenTransaction(ctx context.Context) (tx pgx.Tx, err error)
	Rollback(ctx context.Context, tx pgx.Tx) error
	Commit(ctx context.Context, tx pgx.Tx) error
}
