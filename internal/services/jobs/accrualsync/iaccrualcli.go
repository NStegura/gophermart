package accrualsync

import (
	"context"

	"github.com/NStegura/gophermart/internal/clients/accrual/models"
)

type AccrualCli interface {
	GetOrder(ctx context.Context, number int64) (models.OrderAccrual, error)
}
