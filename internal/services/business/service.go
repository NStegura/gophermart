package business

import (
	"context"
	"errors"
	"fmt"
	"github.com/NStegura/gophermart/internal/customerrors"
	"github.com/NStegura/gophermart/internal/services/business/models"
	"github.com/sirupsen/logrus"
)

type Business struct {
	repo   Repository
	logger *logrus.Logger
}

func New(repo Repository, logger *logrus.Logger) *Business {
	return &Business{repo: repo, logger: logger}
}

func (b *Business) Ping(ctx context.Context) error {
	err := b.repo.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping repo %w", err)
	}
	return nil
}

func (b *Business) CreateUser(ctx context.Context, login, password, salt string) (ID int64, err error) {
	tx, err := b.repo.OpenTransaction(ctx)
	if err != nil {
		return ID, fmt.Errorf("failed to open transaction, %w", err)
	}
	defer func() {
		_ = tx.Commit(ctx)
	}()

	_, err = b.repo.GetUser(ctx, tx, login)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			ID, err = b.repo.CreateUser(ctx, tx, login, password, salt)
			if err != nil {
				return ID, fmt.Errorf("failed to create user, %w", err)
			}
			return
		}
		return ID, fmt.Errorf("failed to get counter metric, %w", err)
	}
	return ID, customerrors.ErrAlreadyExists
}

func (b *Business) GetUser(ctx context.Context, login string) (u models.User, err error) {
	tx, err := b.repo.OpenTransaction(ctx)
	if err != nil {
		return u, fmt.Errorf("failed to open transaction, %w", err)
	}
	defer func() {
		_ = tx.Commit(ctx)
	}()

	dbUser, err := b.repo.GetUser(ctx, tx, login)
	if err != nil {
		return u, fmt.Errorf("failed to create user, %w", err)
	}
	return models.User(dbUser), nil
}
