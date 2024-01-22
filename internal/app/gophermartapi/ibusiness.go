package gophermartapi

import (
	"context"

	domenModels "github.com/NStegura/gophermart/internal/services/business/models"
)

type Business interface {
	Ping(ctx context.Context) error

	CreateUser(ctx context.Context, login, password, salt string) (id int64, err error)
	GetUserByLogin(ctx context.Context, login string) (u domenModels.User, err error)
	GetUserByID(ctx context.Context, ID int64) (u domenModels.User, err error)
	GetOrders(ctx context.Context, userID int64) (orders []domenModels.Order, err error)
	CreateOrder(ctx context.Context, userID int64, orderID int64) error
	CreateWithdraw(ctx context.Context, userID int64, orderID int64, sum float64) error
	GetWithdrawals(ctx context.Context, userID int64) (withdrawals []domenModels.Withdraw, err error)
}
