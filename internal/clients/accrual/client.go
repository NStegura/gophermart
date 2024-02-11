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
	"sync/atomic"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/sirupsen/logrus"

	"github.com/NStegura/gophermart/internal/clients/accrual/models"
)

type Client struct {
	client *http.Client
	logger *logrus.Logger
	URL    string

	isAvailable atomic.Bool
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
	cli := Client{
		client: &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
		URL:    addr,
		logger: logger,
	}
	cli.isAvailable.Store(true)
	return &cli, nil
}

func (c *Client) GetOrder(ctx context.Context, number int64) (models.OrderAccrual, error) {
	var orderAccrual models.OrderAccrual

	if !c.isAvailable.Load() {
		return orderAccrual, ErrClientSemaphore
	}

	reqURL := fmt.Sprintf("%s/api/orders/%v", c.URL, number)
	c.logger.Infof("Get order From accrual client, %s", reqURL)
	resp, err := otelhttp.Get(ctx, reqURL)
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
			c.isAvailable.Store(false)
			c.logger.Debugf("keep client closed for %s seconds", wait)
			time.Sleep(wait)
			c.isAvailable.Store(true)
			c.logger.Debugf("open client")
		}(time.Duration(retryAfter) * time.Second)

		return orderAccrual, ErrTooManyRequests
	}
	return orderAccrual, nil
}
