package business

import (
	"context"
	"github.com/NStegura/gophermart/internal/repo/models"
	"github.com/jackc/pgx/v5"
)

type Repository interface {
	Ping(ctx context.Context) error

	CreateUser(ctx context.Context, tx pgx.Tx, login, password, salt string) (ID int64, err error)
	GetUser(ctx context.Context, tx pgx.Tx, login string) (u models.User, err error)

	OpenTransaction(ctx context.Context) (tx pgx.Tx, err error)
	Rollback(ctx context.Context, tx pgx.Tx) error
	Commit(ctx context.Context, tx pgx.Tx) error
}
