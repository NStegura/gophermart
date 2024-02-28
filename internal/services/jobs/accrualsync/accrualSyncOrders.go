package accrualsync

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/NStegura/gophermart/internal/clients/accrual"
	accrualModels "github.com/NStegura/gophermart/internal/clients/accrual/models"
	"github.com/NStegura/gophermart/internal/repo/models"
)

var (
	ErrNotFoundOrders = errors.New("not orders to sync")
)

type Job struct {
	frequency time.Duration
	rateLimit int

	repo       Repository
	accrualCli AccrualCli
	logger     *logrus.Logger
}

func New(
	frequency time.Duration,
	rateLimit int,
	repo Repository,
	accrualCli *accrual.Client,
	logger *logrus.Logger) *Job {
	return &Job{
		frequency:  frequency,
		rateLimit:  rateLimit,
		repo:       repo,
		accrualCli: accrualCli,
		logger:     logger,
	}
}

func (j *Job) Start(ctx context.Context) error {
	timer := time.NewTicker(j.frequency)
	defer timer.Stop()
	i := 0
	for {
		select {
		case <-timer.C:
			i++
			j.logger.Infof("[JOB|%v] Sync order info", i)
			ordersToSyncCh, err := j.getOrdersToSync(ctx)
			if err != nil {
				j.logger.Errorf("failed to get not processed orders: %s", err)
				continue
			}
			responceCh := make(chan accrualModels.OrderAccrual, len(ordersToSyncCh))

			var wg sync.WaitGroup
			for w := 1; w <= j.rateLimit; w++ {
				wg.Add(1)
				j.getAccrualOrdersResp(ctx, &wg, ordersToSyncCh, responceCh)
			}
			go func() {
				wg.Wait()
				close(responceCh)
			}()

			for respAccrualOrder := range responceCh {
				j.logger.Debugf("Get respAccrualOrder from channel resp %v", respAccrualOrder)
				err = j.updateOrder(ctx, respAccrualOrder)
				if err != nil {
					j.logger.Error(err)
					continue
				}
			}

		case <-ctx.Done():
			return nil
		}
	}
}
func (j *Job) getOrdersToSync(ctx context.Context) (chan models.Order, error) {
	tx, err := j.repo.OpenTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open transaction, %w", err)
	}
	defer func() {
		_ = j.repo.Commit(ctx, tx)
	}()

	ordersToSync, err := j.repo.GetNotProcessedOrders(ctx, tx)
	if len(ordersToSync) == 0 {
		return nil, ErrNotFoundOrders
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get not porocessed orders from db: %w", err)
	}
	ordersToSyncCh := make(chan models.Order, len(ordersToSync))

	go func() {
		defer close(ordersToSyncCh)

		for i := 0; i < len(ordersToSync); i++ {
			j.logger.Debug("get order to sync", ordersToSync[i].ID)
			ordersToSyncCh <- ordersToSync[i]
		}
	}()
	return ordersToSyncCh, nil
}

func (j *Job) getAccrualOrdersResp(
	ctx context.Context,
	wg *sync.WaitGroup,
	ordersToSync chan models.Order,
	responceCh chan accrualModels.OrderAccrual) {
	go func() {
		defer wg.Done()

		for orderToSync := range ordersToSync {
			order, err := j.accrualCli.GetOrder(ctx, orderToSync.ID)
			if err != nil {
				j.logger.Error(err)
			} else {
				responceCh <- order
			}
		}
	}()
}

func (j *Job) updateOrder(ctx context.Context, accrualOrder accrualModels.OrderAccrual) error {
	tx, err := j.repo.OpenTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to open transaction, %w", err)
	}

	order, err := j.repo.GetOrder(ctx, tx, accrualOrder.OrderID, true) // for_update
	if err != nil {
		_ = j.repo.Rollback(ctx, tx)
		return fmt.Errorf("failed to get order for update, %w", err)
	}

	user, err := j.repo.GetUserByID(ctx, tx, order.UserID, true) // for_update
	if err != nil {
		_ = j.repo.Rollback(ctx, tx)
		return fmt.Errorf("failed to get order for update, %w", err)
	}
	if accrualOrder.Status == accrualModels.REGISTERED.String() {
		accrualOrder.Status = models.NEW.String()
	}
	if err = j.repo.UpdateOrder(ctx, tx, order.ID, accrualOrder.Accrual, accrualOrder.Status); err != nil {
		_ = j.repo.Rollback(ctx, tx)
		return fmt.Errorf("failed to update order, %w", err)
	}
	if accrualOrder.Accrual != 0 {
		if err = j.repo.UpdateUserBalance(ctx, tx, user.ID, user.Balance+accrualOrder.Accrual, user.Withdrawn); err != nil {
			_ = j.repo.Rollback(ctx, tx)
			return fmt.Errorf("failed to update user balance, %w", err)
		}
	}

	if err = j.repo.Commit(ctx, tx); err != nil {
		return fmt.Errorf("failef to commit, %w", err)
	}
	return nil
}
