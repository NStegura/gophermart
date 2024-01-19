package gophermartapi

import (
	"context"

	"github.com/NStegura/gophermart/internal/services/business/models"
)

type Business interface {
	Ping(ctx context.Context) error

	CreateUser(ctx context.Context, login, password, salt string) (id int64, err error)
	GetUser(ctx context.Context, login string) (u models.User, err error)

	CreateOrder(ctx context.Context, userID int64, orderUid int64) error
}
