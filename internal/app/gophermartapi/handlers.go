package gophermartapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/NStegura/gophermart/internal/app/gophermartapi/utils"
	domenModels "github.com/NStegura/gophermart/internal/services/business/models"

	"github.com/NStegura/gophermart/internal/app/gophermartapi/models"
	"github.com/NStegura/gophermart/internal/customerrors"
)

const (
	complexityAlgorithm int = 14
)

// register godoc
//
//	@Summary		Register
//	@Description	register
//	@Tags			auth
//	@Accept			json
//	@Param			data	body	models.User	true	"User data"
//	@Success		200
//	@Header			200	{string}	Authorization	"Use this header in other endpoints"
//	@Failure		409
//	@Failure		400
//	@Failure		500
//	@Router			/api/user/register [post]
func (s *APIServer) register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var inputUser models.User
		var token string

		if err := json.NewDecoder(r.Body).Decode(&inputUser); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		newPass, err := s.auth.GeneratePasswordHash(inputUser.Password, complexityAlgorithm)
		if err != nil {
			s.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		uID, err := s.business.CreateUser(r.Context(), inputUser.Login, newPass)
		if err != nil {
			if errors.Is(err, customerrors.ErrAlreadyExists) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			s.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		token, err = s.auth.GenerateToken(uID)
		if err != nil {
			s.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Authorization", token)
		w.WriteHeader(http.StatusOK)
	}
}

// login godoc
//
//	@Summary		Login
//	@Description	login
//	@Tags			auth
//	@Accept			json
//	@Param			data	body	models.User	true	"User data"
//	@Success		200
//	@Header			200	{string}	Authorization	"Use this header in other endpoints"
//	@Failure		400
//	@Failure		401
//	@Failure		500
//	@Router			/api/user/login [post]
func (s *APIServer) login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var inputUser models.User
		var token string

		if err := json.NewDecoder(r.Body).Decode(&inputUser); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dbUser, err := s.business.GetUserByLogin(r.Context(), inputUser.Login)
		if err != nil {
			if errors.Is(err, errors.New("doesnt exist")) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		if !s.auth.CheckPasswordHash(inputUser.Password, dbUser.Password) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		token, err = s.auth.GenerateToken(dbUser.ID)
		if err != nil {
			s.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Authorization", token)
		w.WriteHeader(http.StatusOK)
	}
}

// createOrder godoc
//
//	@Summary		Create order
//	@Description	"register" user order
//	@Tags			user
//	@Accept			plain
//	@Param			string	body	string	true	"Order id"
//	@Success		200
//	@Success		202
//	@Failure		400
//	@Failure		401
//	@Failure		409
//	@Failure		422
//	@Failure		500
//	@Security		ApiKeyAuth
//	@Router			/api/user/orders [post]
func (s *APIServer) createOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			orderUID int64
			data     []byte
		)

		userID, err := s.getUserID(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		data, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			s.logger.Error(err)
			return
		}
		defer func() {
			_ = r.Body.Close()
		}()

		orderUID, err = strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !utils.Valid(orderUID) {
			http.Error(w, "invalid order format", http.StatusUnprocessableEntity)
			return
		}

		if err = s.business.CreateOrder(r.Context(), userID, orderUID); err != nil {
			switch {
			case errors.Is(err, customerrors.ErrCurrUserUploaded):
				w.WriteHeader(http.StatusOK)
				return
			case errors.Is(err, customerrors.ErrAnotherUserUploaded):
				w.WriteHeader(http.StatusConflict)
				return
			default:
				s.logger.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

// getOrderList godoc
//
//	@Summary		Get order list
//	@Description	get order list by user
//	@Tags			user
//	@Produce		json
//	@Success		200	{array}	models.Order
//	@Failure		204
//	@Failure		401
//	@Failure		500
//	@Security		ApiKeyAuth
//	@Router			/api/user/orders [get]
func (s *APIServer) getOrderList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var orders []models.Order
		var domenOrders []domenModels.Order

		userID, err := s.getUserID(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		domenOrders, err = s.business.GetOrders(r.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, o := range domenOrders {
			orders = append(orders, models.Order(o))
		}

		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		s.writeJSONResp(orders, w)
	}
}

// getBalance godoc
//
//	@Summary		Get balance
//	@Description	get user balance
//	@Tags			user
//	@Produce		json
//	@Success		200	{object}	models.Balance
//	@Failure		401
//	@Failure		500
//	@Security		ApiKeyAuth
//	@Router			/api/user/balance [get]
func (s *APIServer) getBalance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var domenUser domenModels.User

		userID, err := s.getUserID(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		domenUser, err = s.business.GetUserByID(r.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.writeJSONResp(models.Balance{
			Current:   domenUser.Balance,
			Withdrawn: domenUser.Withdrawn,
		}, w)
	}
}

// createWithdraw godoc
//
//	@Summary		Create withdraw
//	@Description	create user withdraw
//	@Tags			user
//	@Accept			json
//	@Param			data	body	models.WithdrawIn	true	"User withdraw data"
//	@Success		200
//	@Failure		401
//	@Failure		402
//	@Failure		422
//	@Failure		500
//	@Security		ApiKeyAuth
//	@Router			/api/user/balance/withdraw [post]
func (s *APIServer) createWithdraw() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			orderUID int64
			withdraw models.WithdrawIn
		)

		userID, err := s.getUserID(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err = json.NewDecoder(r.Body).Decode(&withdraw); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		orderUID, err = strconv.ParseInt(withdraw.Order, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !utils.Valid(orderUID) {
			http.Error(w, "invalid order format", http.StatusUnprocessableEntity)
			return
		}

		if err = s.business.CreateWithdraw(r.Context(), userID, orderUID, withdraw.Sum); err != nil {
			if errors.Is(err, customerrors.ErrNotEnoughFunds) {
				w.WriteHeader(http.StatusPaymentRequired)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// getWithdrawals godoc
//
//	@Summary		Get withdraw list
//	@Description	get user withdraw list
//	@Tags			user
//	@Produce		json
//	@Success		200	{array}	models.WithdrawOut
//	@Failure		401
//	@Failure		500
//	@Security		ApiKeyAuth
//	@Router			/api/user/withdrawals [get]
func (s *APIServer) getWithdrawals() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var withdrawals []models.WithdrawOut
		userID, err := s.getUserID(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		domenWithdrawals, err := s.business.GetWithdrawals(r.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, wd := range domenWithdrawals {
			withdrawals = append(withdrawals, models.WithdrawOut{
				Order:       strconv.FormatInt(wd.OrderID, 10),
				Sum:         wd.Sum,
				ProcessedAt: wd.CreatedAt,
			})
		}

		if len(withdrawals) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		s.writeJSONResp(withdrawals, w)
	}
}

// ToDo: add pagination.
func (s *APIServer) getOrderPaginateList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}

// ping godoc
//
//	@Summary		Get ping
//	@Description	check service
//	@Tags			tech
//	@Success		200
//	@Router			/ping [get]
func (s *APIServer) ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.business.Ping(r.Context()); err != nil {
			s.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
