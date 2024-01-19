package business

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/NStegura/gophermart/internal/customerrors"
	dbModels "github.com/NStegura/gophermart/internal/repo/models"
	domenModels "github.com/NStegura/gophermart/internal/services/business/models"
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

func (b *Business) CreateUser(ctx context.Context, login, password, salt string) (id int64, err error) {
	tx, err := b.repo.OpenTransaction(ctx)
	if err != nil {
		return id, fmt.Errorf("failed to open transaction, %w", err)
	}
	defer func() {
		_ = tx.Commit(ctx)
	}()

	_, err = b.repo.GetUser(ctx, tx, login)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			id, err = b.repo.CreateUser(ctx, tx, login, password, salt)
			if err != nil {
				return id, fmt.Errorf("failed to create user, %w", err)
			}
			return
		}
		return id, fmt.Errorf("failed to get user, %w", err)
	}
	return id, customerrors.ErrAlreadyExists
}

func (b *Business) GetUser(ctx context.Context, login string) (u domenModels.User, err error) {
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
	return domenModels.User(dbUser), nil
}

func (b *Business) CreateOrder(ctx context.Context, userID, orderID int64) error {
	var dbOrder dbModels.Order

	tx, err := b.repo.OpenTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to open transaction, %w", err)
	}
	defer func() {
		_ = tx.Commit(ctx)
	}()

	dbOrder, err = b.repo.GetOrder(ctx, tx, orderID)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			err = b.repo.CreateOrder(ctx, tx, userID, orderID)
			if err != nil {
				return fmt.Errorf("failed to create user, %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to get user, %w", err)
	}
	if dbOrder.UserID != userID {
		return customerrors.ErrAnotherUserUploaded
	}
	return customerrors.ErrCurrUserUploaded
}
