package business

import (
	"context"
	"errors"
	"fmt"
	"time"

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
		_ = b.repo.Commit(ctx, tx)
	}()

	_, err = b.repo.GetUserByLogin(ctx, tx, login)
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

func (b *Business) GetUserByLogin(ctx context.Context, login string) (u domenModels.User, err error) {
	tx, err := b.repo.OpenTransaction(ctx)
	if err != nil {
		return u, fmt.Errorf("failed to open transaction, %w", err)
	}
	defer func() {
		_ = b.repo.Commit(ctx, tx)
	}()

	dbUser, err := b.repo.GetUserByLogin(ctx, tx, login)
	if err != nil {
		return u, fmt.Errorf("failed to get user, %w", err)
	}
	return domenModels.User(dbUser), nil
}

func (b *Business) GetUserByID(ctx context.Context, id int64) (u domenModels.User, err error) {
	tx, err := b.repo.OpenTransaction(ctx)
	if err != nil {
		return u, fmt.Errorf("failed to open transaction, %w", err)
	}
	defer func() {
		_ = b.repo.Commit(ctx, tx)
	}()

	dbUser, err := b.repo.GetUserByID(ctx, tx, id, false)
	if err != nil {
		return u, fmt.Errorf("failed to get user, %w", err)
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
		_ = b.repo.Commit(ctx, tx)
	}()

	dbOrder, err = b.repo.GetOrder(ctx, tx, orderID, false)
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

func (b *Business) GetOrders(ctx context.Context, userID int64) (orders []domenModels.Order, err error) {
	tx, err := b.repo.OpenTransaction(ctx)
	if err != nil {
		return orders, fmt.Errorf("failed to open transaction, %w", err)
	}
	defer func() {
		_ = b.repo.Commit(ctx, tx)
	}()

	dbOrders, err := b.repo.GetOrders(ctx, tx, userID)
	if err != nil {
		return orders, fmt.Errorf("failed to get orders, %w", err)
	}
	for _, dbOrder := range dbOrders {
		var convertedUpdatedAt time.Time
		convertedUpdatedAt, err = time.Parse(time.RFC3339, dbOrder.UpdatedAt.Format(time.RFC3339))
		if err != nil {
			return orders, fmt.Errorf("failed to convert UpdatedAt to RFC3339")
		}
		orders = append(orders, domenModels.Order{
			Number:     dbOrder.ID,
			Status:     dbOrder.Status,
			Accrual:    dbOrder.Accrual.Float64,
			UploadedAt: convertedUpdatedAt,
		})
	}
	return orders, nil
}

func (b *Business) CreateWithdraw(ctx context.Context, userID int64, orderID int64, sum float64) error {
	tx, err := b.repo.OpenTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to open transaction, %w", err)
	}

	user, err := b.repo.GetUserByID(ctx, tx, userID, true)
	if err != nil {
		_ = b.repo.Rollback(ctx, tx)
		return fmt.Errorf("failed to get user, %w", err)
	}
	if user.Balance < sum {
		_ = b.repo.Rollback(ctx, tx)
		return customerrors.ErrNotEnoughFunds
	}

	err = b.repo.UpdateUserBalance(ctx, tx, user.ID, user.Balance-sum, user.Withdrawn+sum)
	if err != nil {
		_ = b.repo.Rollback(ctx, tx)
		return fmt.Errorf("failed to create withdraw, %w", err)
	}

	err = b.repo.CreateWithdraw(ctx, tx, userID, orderID, sum)
	if err != nil {
		_ = b.repo.Rollback(ctx, tx)
		return fmt.Errorf("failed to create withdraw, %w", err)
	}
	_ = b.repo.Commit(ctx, tx)
	return nil
}

func (b *Business) GetWithdrawals(ctx context.Context, userID int64) (withdrawals []domenModels.Withdraw, err error) {
	tx, err := b.repo.OpenTransaction(ctx)
	if err != nil {
		return withdrawals, fmt.Errorf("failed to open transaction, %w", err)
	}
	defer func() {
		_ = b.repo.Commit(ctx, tx)
	}()

	dbWithdrawals, err := b.repo.GetWithdrawals(ctx, tx, userID)
	if err != nil {
		return withdrawals, fmt.Errorf("failed to get withdrawals, %w", err)
	}
	for _, dbWithdraw := range dbWithdrawals {
		var convertedCreatedAt time.Time
		convertedCreatedAt, err = time.Parse(time.RFC3339, dbWithdraw.CreatedAt.Format(time.RFC3339))
		if err != nil {
			return withdrawals, fmt.Errorf("failed to convert UpdatedAt to RFC3339")
		}
		withdrawals = append(withdrawals, domenModels.Withdraw{
			OrderID:   dbWithdraw.OrderID,
			Sum:       dbWithdraw.Sum,
			CreatedAt: convertedCreatedAt,
		})
	}
	return withdrawals, nil
}
