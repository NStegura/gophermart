package gophermartapi

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/NStegura/gophermart/internal/app/gophermartapi/models"
	"github.com/NStegura/gophermart/internal/customerrors"
	"github.com/NStegura/gophermart/internal/monitoring/logger"
	mock_gophermartapi "github.com/NStegura/gophermart/mocks/app/gophermartapi"
)

type testHelper struct {
	ctrl         *gomock.Controller
	ts           *httptest.Server
	mockBusiness *mock_gophermartapi.MockBusiness
	mockAuth     *mock_gophermartapi.MockAuth
}

func (th *testHelper) request(t *testing.T, method, path string, body io.Reader) (map[string][]string, int, string) {
	t.Helper()
	req, err := http.NewRequest(method, th.ts.URL+path, body)
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
				th.mockAuth.EXPECT().GeneratePasswordHash(gomock.Any(), gomock.Any()).Return("newPass", nil),
				th.mockBusiness.EXPECT().CreateUser(gomock.Any(), test.inputUser.Login, "newPass").Return(int64(1), test.err),
				th.mockAuth.EXPECT().GenerateToken(gomock.Any()).Return("token", nil))

			headers, statusCode, respBodyStr := th.request(t, "POST", "/api/user/register",
				bytes.NewBufferString(test.inputBody))
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
		name                 string
		inputBody            string
		inputUser            models.User
		err                  error
		expectedStatusCode   int
		requireGenerateToken bool
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
				bytes.NewBufferString(test.inputBody))

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
			name:      "Ok",
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
				bytes.NewBufferString(test.inputBody))

			// require
			require.Equal(t, statusCode, test.expectedStatusCode)
		})
	}
}
