package gophermartapi

import (
	"context"
	"github.com/NStegura/gophermart/internal/services/business/models"
)

type Business interface {
	Ping(ctx context.Context) error

	CreateUser(ctx context.Context, login, password, salt string) (ID int64, err error)
	GetUser(ctx context.Context, login string) (u models.User, err error)
}
