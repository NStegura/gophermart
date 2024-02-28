package accrual

import (
	"errors"
)

var (
	ErrClientSemaphore     = errors.New("client closed by semaphore")
	ErrInvalidOrderAccrual = errors.New("invalid orderAccrual")
	ErrNoContent           = errors.New("no content")
	ErrTooManyRequests     = errors.New("too many requests")
)
