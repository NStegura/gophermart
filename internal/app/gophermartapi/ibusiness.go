package gophermartapi

import (
	"context"

	domenModels "github.com/NStegura/gophermart/internal/services/business/models"
)

type Business interface {
	Ping(ctx context.Context) error

	CreateUser(ctx context.Context, login, password, salt string) (id int64, err error)
	GetUser(ctx context.Context, login string) (u domenModels.User, err error)
	GetOrders(ctx context.Context, userID int64) (orders []domenModels.Order, err error)
	CreateOrder(ctx context.Context, userID int64, orderUid int64) error
}
