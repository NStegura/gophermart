package gophermartapi

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/NStegura/gophermart/internal/app/gophermartapi/models"
	"github.com/NStegura/gophermart/internal/customerrors"
	"github.com/NStegura/gophermart/internal/monitoring/logger"
	domenModels "github.com/NStegura/gophermart/internal/services/business/models"
	mock_gophermartapi "github.com/NStegura/gophermart/mocks/app/gophermartapi"
)

type testHelper struct {
	ctrl         *gomock.Controller
	ts           *httptest.Server
	mockBusiness *mock_gophermartapi.MockBusiness
	mockAuth     *mock_gophermartapi.MockAuth
}

func (th *testHelper) request(
	t *testing.T,
	method, path string,
	body io.Reader,
	headers *map[string]string,
) (map[string][]string, int, string) {
	t.Helper()
	req, err := http.NewRequest(method, th.ts.URL+path, body)
	if headers != nil {
		for k, v := range *headers {
			req.Header.Add(k, v)
		}
	}
	require.NoError(t, err)

	resp, err := th.ts.Client().Do(req)
	require.NoError(t, err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Log(err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.Header, resp.StatusCode, string(respBody)
}

func initTestHelper(t *testing.T) *testHelper {
	t.Helper()
	ctrl := gomock.NewController(t)
	cfglog, _ := logger.Init("info")
	mockBusiness := mock_gophermartapi.NewMockBusiness(ctrl)
	mockAuth := mock_gophermartapi.NewMockAuth(ctrl)

	server := New(
		":8080",
		mockBusiness,
		mockAuth,
		cfglog,
	)
	server.configRouter()

	ts := httptest.NewServer(server.router)
	return &testHelper{
		ctrl:         ctrl,
		ts:           ts,
		mockBusiness: mockBusiness,
		mockAuth:     mockAuth,
	}
}

func (th *testHelper) finish() {
	th.ts.Close()
	th.ctrl.Finish()
}

func TestHandler_Register__ok(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name               string
		inputBody          string
		inputUser          models.User
		err                error
		expectedStatusCode int
	}{
		{
			name:      "Ok",
			inputBody: `{"login": "login", "password": "password"}`,
			inputUser: models.User{
				Login:    "login",
				Password: "password",
			},
			err:                nil,
			expectedStatusCode: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockAuth.EXPECT().GeneratePasswordHash(gomock.Any(), gomock.Any()).Return("newPass", nil),
				th.mockBusiness.EXPECT().CreateUser(gomock.Any(), test.inputUser.Login, "newPass").Return(int64(1), test.err),
				th.mockAuth.EXPECT().GenerateToken(gomock.Any()).Return("token", nil))

			headers, statusCode, respBodyStr := th.request(t, "POST", "/api/user/register",
				bytes.NewBufferString(test.inputBody), nil)
			t.Log(respBodyStr)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
			require.Equal(t, headers["Authorization"][0], "token")
		})
	}
}

func TestHandler_Register__bad_resp(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name               string
		inputBody          string
		inputUser          models.User
		err                error
		expectedStatusCode int
	}{
		{
			name:      "Wrong Input",
			inputBody: `{sdfsdf}`,
			inputUser: models.User{
				Login:    "username",
				Password: "qwerty",
			},
			err:                nil,
			expectedStatusCode: 400,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, statusCode, _ := th.request(t, "POST", "/api/user/register",
				bytes.NewBufferString(test.inputBody), nil)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_Register__userAlreadyExists(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name                 string
		inputBody            string
		inputUser            models.User
		err                  error
		expectedStatusCode   int
		requireGenerateToken bool
	}{
		{
			name:      "userAlreadyExists",
			inputBody: `{"login": "login", "password": "password"}`,
			inputUser: models.User{
				Login:    "login",
				Password: "password",
			},
			err:                customerrors.ErrAlreadyExists,
			expectedStatusCode: 409,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockAuth.EXPECT().GeneratePasswordHash(gomock.Any(), gomock.Any()).Return("newPass", nil),
				th.mockBusiness.EXPECT().CreateUser(gomock.Any(), test.inputUser.Login, "newPass").Return(int64(1), test.err),
			)

			_, statusCode, _ := th.request(t, "POST", "/api/user/register",
				bytes.NewBufferString(test.inputBody), nil)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_Login__Ok(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name                 string
		inputBody            string
		inputUser            models.User
		err                  error
		expectedStatusCode   int
		requireGenerateToken bool
	}{
		{
			name:      "Ok",
			inputBody: `{"login": "login", "password": "password"}`,
			inputUser: models.User{
				Login:    "login",
				Password: "password",
			},
			err:                nil,
			expectedStatusCode: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockBusiness.EXPECT().GetUserByLogin(gomock.Any(), test.inputUser.Login).Return(domenModels.User{
					ID: 1, Login: test.inputUser.Login, Password: "hash_pass", Balance: 100, Withdrawn: 10, CreatedAt: time.Now(),
				}, nil),
				th.mockAuth.EXPECT().CheckPasswordHash(gomock.Any(), gomock.Any()).Return(true),
				th.mockAuth.EXPECT().GenerateToken(gomock.Any()).Return("token", nil),
			)

			_, statusCode, _ := th.request(t, "POST", "/api/user/login",
				bytes.NewBufferString(test.inputBody), nil)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_Login__Unauthorized(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name                 string
		inputBody            string
		inputUser            models.User
		err                  error
		expectedStatusCode   int
		requireGenerateToken bool
	}{
		{
			name:      "Unauthorized",
			inputBody: `{"login": "login", "password": "password"}`,
			inputUser: models.User{
				Login:    "login",
				Password: "password",
			},
			err:                nil,
			expectedStatusCode: 401,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockBusiness.EXPECT().GetUserByLogin(gomock.Any(), test.inputUser.Login).Return(domenModels.User{
					ID: 1, Login: test.inputUser.Login, Password: "hash_pass", Balance: 100, Withdrawn: 10, CreatedAt: time.Now(),
				}, nil),
				th.mockAuth.EXPECT().CheckPasswordHash(gomock.Any(), gomock.Any()).Return(false),
			)

			_, statusCode, _ := th.request(t, "POST", "/api/user/login",
				bytes.NewBufferString(test.inputBody), nil)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_Login__BadRequest(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name               string
		inputBody          string
		err                error
		expectedStatusCode int
	}{
		{
			name:               "BadRequest",
			inputBody:          ``,
			err:                nil,
			expectedStatusCode: 400,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, statusCode, _ := th.request(t, "POST", "/api/user/login",
				bytes.NewBufferString(test.inputBody), nil)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_createOrder__Ok(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name               string
		inputBody          string
		err                error
		expectedStatusCode int
	}{
		{
			name:               "Ok",
			inputBody:          `1234567897`,
			err:                nil,
			expectedStatusCode: 202,
		},
	}

	headers := make(map[string]string, 1)
	headers["Authorization"] = "auth header"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockAuth.EXPECT().ParseToken(gomock.Any()).Return(int64(1), nil),
				th.mockBusiness.EXPECT().CreateOrder(gomock.Any(), int64(1), int64(1234567897)).Return(nil),
			)
			_, statusCode, _ := th.request(t, "POST", "/api/user/orders",
				bytes.NewBufferString(test.inputBody), &headers)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_createOrder__Unauthorized_NotHeader(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name               string
		inputBody          string
		err                error
		expectedStatusCode int
	}{
		{
			name:               "Unauthorized_NotHeader",
			inputBody:          `1234567897`,
			err:                nil,
			expectedStatusCode: 401,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, statusCode, _ := th.request(t, "POST", "/api/user/orders",
				bytes.NewBufferString(test.inputBody), nil)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_createOrder__NotValidOrderNum(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name               string
		inputBody          string
		err                error
		expectedStatusCode int
	}{
		{
			name:               "NotValidOrderNum",
			inputBody:          `1`,
			err:                nil,
			expectedStatusCode: 422,
		},
	}

	headers := make(map[string]string, 1)
	headers["Authorization"] = "auth header"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockAuth.EXPECT().ParseToken(gomock.Any()).Return(int64(1), nil),
			)
			_, statusCode, _ := th.request(t, "POST", "/api/user/orders",
				bytes.NewBufferString(test.inputBody), &headers)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_createOrder__AlreadyUploaded(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name               string
		inputBody          string
		err                error
		expectedStatusCode int
	}{
		{
			name:               "AlreadyUploaded",
			inputBody:          `1234567897`,
			err:                customerrors.ErrCurrUserUploaded,
			expectedStatusCode: 200,
		},
	}

	headers := make(map[string]string, 1)
	headers["Authorization"] = "auth header"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockAuth.EXPECT().ParseToken(gomock.Any()).Return(int64(1), nil),
				th.mockBusiness.EXPECT().CreateOrder(gomock.Any(), int64(1), int64(1234567897)).Return(test.err),
			)
			_, statusCode, _ := th.request(t, "POST", "/api/user/orders",
				bytes.NewBufferString(test.inputBody), &headers)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_createOrder__AnotherUserUploaded(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name               string
		inputBody          string
		err                error
		expectedStatusCode int
	}{
		{
			name:               "AlreadyUploaded",
			inputBody:          `1234567897`,
			err:                customerrors.ErrAnotherUserUploaded,
			expectedStatusCode: 409,
		},
	}

	headers := make(map[string]string, 1)
	headers["Authorization"] = "auth header"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockAuth.EXPECT().ParseToken(gomock.Any()).Return(int64(1), nil),
				th.mockBusiness.EXPECT().CreateOrder(gomock.Any(), int64(1), int64(1234567897)).Return(test.err),
			)
			_, statusCode, _ := th.request(t, "POST", "/api/user/orders",
				bytes.NewBufferString(test.inputBody), &headers)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_createOrder__UnexpectedErr(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name               string
		inputBody          string
		err                error
		expectedStatusCode int
	}{
		{
			name:               "AlreadyUploaded",
			inputBody:          `1234567897`,
			err:                errors.New("some err"),
			expectedStatusCode: 500,
		},
	}

	headers := make(map[string]string, 1)
	headers["Authorization"] = "auth header"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockAuth.EXPECT().ParseToken(gomock.Any()).Return(int64(1), nil),
				th.mockBusiness.EXPECT().CreateOrder(gomock.Any(), int64(1), int64(1234567897)).Return(test.err),
			)
			_, statusCode, _ := th.request(t, "POST", "/api/user/orders",
				bytes.NewBufferString(test.inputBody), &headers)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_getOrderList__Ok(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name                 string
		err                  error
		expectedStatusCode   int
		requireGenerateToken bool
	}{
		{
			name:               "ok",
			err:                nil,
			expectedStatusCode: 200,
		},
	}

	headers := make(map[string]string, 1)
	headers["Authorization"] = "auth header"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockAuth.EXPECT().ParseToken(gomock.Any()).Return(int64(1), nil),
				th.mockBusiness.EXPECT().GetOrders(gomock.Any(), int64(1)).Return([]domenModels.Order{{}}, nil),
			)
			_, statusCode, _ := th.request(t, "GET", "/api/user/orders",
				bytes.NewBufferString(``), &headers)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_getOrderList__NoContent(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name                 string
		err                  error
		expectedStatusCode   int
		requireGenerateToken bool
	}{
		{
			name:               "ok",
			err:                nil,
			expectedStatusCode: 204,
		},
	}

	headers := make(map[string]string, 1)
	headers["Authorization"] = "auth header"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockAuth.EXPECT().ParseToken(gomock.Any()).Return(int64(1), nil),
				th.mockBusiness.EXPECT().GetOrders(gomock.Any(), int64(1)).Return([]domenModels.Order{}, nil),
			)
			_, statusCode, _ := th.request(t, "GET", "/api/user/orders",
				bytes.NewBufferString(``), &headers)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_getBalance__Ok(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name                 string
		err                  error
		expectedStatusCode   int
		requireGenerateToken bool
	}{
		{
			name:               "ok",
			err:                nil,
			expectedStatusCode: 200,
		},
	}

	headers := make(map[string]string, 1)
	headers["Authorization"] = "auth header"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockAuth.EXPECT().ParseToken(gomock.Any()).Return(int64(1), nil),
				th.mockBusiness.EXPECT().GetUserByID(gomock.Any(), int64(1)).Return(domenModels.User{}, nil),
			)
			_, statusCode, _ := th.request(t, "GET", "/api/user/balance",
				bytes.NewBufferString(``), &headers)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_createWithdraw__Ok(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name                 string
		inputBody            string
		err                  error
		expectedStatusCode   int
		requireGenerateToken bool
	}{
		{
			name:               "ok",
			inputBody:          `{"order": "1234567897", "sum": 50}`,
			err:                nil,
			expectedStatusCode: 200,
		},
	}

	headers := make(map[string]string, 1)
	headers["Authorization"] = "auth header"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockAuth.EXPECT().ParseToken(gomock.Any()).Return(int64(1), nil),
				th.mockBusiness.EXPECT().CreateWithdraw(gomock.Any(), int64(1), int64(1234567897), float64(50)).Return(nil),
			)
			_, statusCode, _ := th.request(t, "POST", "/api/user/balance/withdraw",
				bytes.NewBufferString(test.inputBody), &headers)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_createWithdraw__NotValidOrderNum(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name                 string
		inputBody            string
		err                  error
		expectedStatusCode   int
		requireGenerateToken bool
	}{
		{
			name:               "NotValidOrderNum",
			inputBody:          `{"order": "1", "sum": 50}`,
			err:                nil,
			expectedStatusCode: 422,
		},
	}

	headers := make(map[string]string, 1)
	headers["Authorization"] = "auth header"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockAuth.EXPECT().ParseToken(gomock.Any()).Return(int64(1), nil),
			)
			_, statusCode, _ := th.request(t, "POST", "/api/user/balance/withdraw",
				bytes.NewBufferString(test.inputBody), &headers)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_createWithdraw__NotEnoughFunds(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name                 string
		inputBody            string
		err                  error
		expectedStatusCode   int
		requireGenerateToken bool
	}{
		{
			name:               "NotEnoughFunds",
			inputBody:          `{"order": "1234567897", "sum": 50}`,
			err:                customerrors.ErrNotEnoughFunds,
			expectedStatusCode: 402,
		},
	}

	headers := make(map[string]string, 1)
	headers["Authorization"] = "auth header"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockAuth.EXPECT().ParseToken(gomock.Any()).Return(int64(1), nil),
				th.mockBusiness.EXPECT().CreateWithdraw(gomock.Any(), int64(1), int64(1234567897), float64(50)).Return(test.err),
			)
			_, statusCode, _ := th.request(t, "POST", "/api/user/balance/withdraw",
				bytes.NewBufferString(test.inputBody), &headers)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}

func TestHandler_getWithdrawals__Ok(t *testing.T) {
	th := initTestHelper(t)
	defer th.finish()

	tests := []struct {
		name                 string
		err                  error
		expectedStatusCode   int
		requireGenerateToken bool
	}{
		{
			name:               "Ok",
			err:                nil,
			expectedStatusCode: 200,
		},
	}

	headers := make(map[string]string, 1)
	headers["Authorization"] = "auth header"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gomock.InOrder(
				th.mockAuth.EXPECT().ParseToken(gomock.Any()).Return(int64(1), nil),
				th.mockBusiness.EXPECT().GetWithdrawals(gomock.Any(), int64(1)).Return([]domenModels.Withdraw{{}}, nil),
			)
			_, statusCode, _ := th.request(t, "GET", "/api/user/withdrawals",
				bytes.NewBufferString(``), &headers)

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}
