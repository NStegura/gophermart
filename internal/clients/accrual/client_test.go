package accrual

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/NStegura/gophermart/internal/clients/accrual/models"
)

type MockHTTPCLient struct {
	expectedResponse *http.Response
	expectedErr      error
}

func (mock *MockHTTPCLient) Get(_ string) (*http.Response, error) {
	return mock.expectedResponse, mock.expectedErr
}

func (mock *MockHTTPCLient) Expect(exResp *http.Response, exErr error) {
	mock.expectedResponse = exResp
	mock.expectedErr = exErr
}

func TestServiceAccrual_ProcessedAccrualData(t *testing.T) {
	var (
		emptyOrderAccrual models.OrderAccrual
		testOrderNum      int64 = 371449635398431
		someErr                 = errors.New("some error")
		badResp                 = fmt.Sprintf("{\"order\":\"%v\",\"status\":\"SomeStatus\",\"accrual\":500}", testOrderNum)
		goodResp                = fmt.Sprintf("{\"order\":\"%v\",\"status\":\"PROCESSED\",\"accrual\":500}", testOrderNum)
	)

	testAddr := "http://testhost:8090"

	type cliMock struct {
		resp *http.Response
		err  error
	}

	type expected struct {
		orderAccrual models.OrderAccrual
		err          error
	}

	tests := []struct {
		name string
		cliMock
		expected
	}{
		{
			"Successful Response",
			cliMock{&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(goodResp))),
			}, nil},
			expected{
				models.OrderAccrual{
					OrderID: 371449635398431,
					Status:  models.PROCESSED.String(),
					Accrual: 500}, nil},
		},
		{
			"Err from Response",
			cliMock{nil, someErr},
			expected{emptyOrderAccrual, someErr},
		},
		{
			"Bad Response",
			cliMock{&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(badResp))),
			}, nil},
			expected{emptyOrderAccrual, ErrInvalidOrderAccrual},
		},
		{
			"StatusNoContent",
			cliMock{&http.Response{StatusCode: http.StatusNoContent,
				Body: io.NopCloser(bytes.NewReader([]byte(badResp)))}, nil},
			expected{emptyOrderAccrual, ErrNoContent},
		},
		{
			"StatusTooManyRequests",
			cliMock{&http.Response{StatusCode: http.StatusTooManyRequests,
				Body: io.NopCloser(bytes.NewReader([]byte(badResp)))}, nil},
			expected{emptyOrderAccrual, ErrTooManyRequests},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accrualCli, _ := New(testAddr, logrus.New())
			mockCli := MockHTTPCLient{}
			accrualCli.client = &mockCli
			accrualCli.collectMetrics = false

			mockCli.Expect(tt.cliMock.resp, tt.cliMock.err)

			ctx := context.Background()
			order, err := accrualCli.GetOrder(ctx, testOrderNum)
			if err != nil {
				assert.Equal(t, true, errors.Is(err, tt.expected.err))
			} else {
				assert.Equal(t, tt.expected.orderAccrual, order)
			}
		})
	}
}
