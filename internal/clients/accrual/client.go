package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	"github.com/NStegura/gophermart/internal/clients/accrual/models"
)

type Client struct {
	client *http.Client
	logger *logrus.Logger
	URL    string

	semaphore *semaphore.Weighted
}

func New(
	addr string,
	logger *logrus.Logger,
) (*Client, error) {
	var err error
	if !strings.HasPrefix(addr, "http") {
		addr, err = url.JoinPath("http:", addr)
		if err != nil {
			return nil, fmt.Errorf("failed to init client, %w", err)
		}
	}
	return &Client{
		client:    &http.Client{},
		URL:       addr,
		logger:    logger,
		semaphore: semaphore.NewWeighted(1),
	}, nil
}

type RequestError struct {
	URL        *url.URL
	Body       []byte
	StatusCode int
}

func (e RequestError) Error() string {
	return fmt.Sprintf(
		"Metric request error: url=%s, code=%v, body=%s",
		e.URL, e.StatusCode, e.Body,
	)
}

func (c *Client) GetOrder(ctx context.Context, number int64) (models.OrderAccrual, error) {
	var orderAccrual models.OrderAccrual

	ok := c.semaphore.TryAcquire(1)
	if !ok {
		return orderAccrual, ErrClientSemaphore
	} else {
		c.semaphore.Release(1)
	}

	reqURL := fmt.Sprintf("%s/api/orders/%v", c.URL, number)
	c.logger.Infof("Get order From accrual client, %s", reqURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		c.logger.Error(err)
		return orderAccrual, fmt.Errorf("failed to make request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("Can't get resp orderAccrual")
		return orderAccrual, fmt.Errorf("failed to get resp orderAccrual: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			c.logger.Error(err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		jsonData, err := io.ReadAll(resp.Body)
		if err != nil {
			c.logger.Error(err)
			return orderAccrual, fmt.Errorf("failed to read body: %w", err)
		}

		if err := json.Unmarshal(jsonData, &orderAccrual); err != nil {
			c.logger.Error(err)
			return orderAccrual, fmt.Errorf("failed to unmarshal responce: %w", err)
		}
		if !orderAccrual.IsValid() {
			c.logger.Error("Invalid order Accrual response", orderAccrual)
			return orderAccrual, ErrInvalidOrderAccrual
		}
		c.logger.Debug("Get data orderAccrual accrual", orderAccrual)
		return orderAccrual, nil

	case http.StatusNoContent:
		c.logger.Info("No content in request ")
		return orderAccrual, ErrNoContent

	case http.StatusTooManyRequests:
		c.logger.Info("Too Many Requests ")

		retryHeader := resp.Header.Get("Retry-After")
		retryAfter, err := strconv.Atoi(retryHeader)
		if err != nil {
			return orderAccrual, ErrTooManyRequests
		}

		go func(wait time.Duration) {
			if err = c.semaphore.Acquire(ctx, 1); err != nil {
				c.logger.Error("failed to Acquire semaphore")
				return
			}
			c.logger.Debugf("keep semaphore closed for %s seconds", wait)
			time.Sleep(wait)
			c.semaphore.Release(1)
		}(time.Duration(retryAfter) * time.Second)

		return orderAccrual, ErrTooManyRequests
	}
	return orderAccrual, nil
}
