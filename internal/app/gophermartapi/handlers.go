package gophermartapi

import (
	"encoding/json"
	"errors"
	"github.com/NStegura/gophermart/internal/app/gophermartapi/utils"
	domenModels "github.com/NStegura/gophermart/internal/services/business/models"
	"io"
	"net/http"
	"strconv"

	"github.com/NStegura/gophermart/internal/app/gophermartapi/models"
	"github.com/NStegura/gophermart/internal/customerrors"
)

const (
	complexityAlgorithm int64 = 30
)

func (s *APIServer) register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var inputUser models.User
		var token string

		if err := json.NewDecoder(r.Body).Decode(&inputUser); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		salt := s.auth.GenerateUserSalt(complexityAlgorithm)
		newPass := s.auth.GeneratePasswordHash(inputUser.Password, salt)

		uID, err := s.business.CreateUser(r.Context(), inputUser.Login, newPass, salt)
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

func (s *APIServer) login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var inputUser models.User
		var token string

		if err := json.NewDecoder(r.Body).Decode(&inputUser); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dbUser, err := s.business.GetUser(r.Context(), inputUser.Login)
		if err != nil {
			if errors.Is(err, errors.New("doesnt exist")) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		if dbUser.Password != s.auth.GeneratePasswordHash(inputUser.Password, dbUser.Salt) {
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

func (s *APIServer) createOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var orderUid int64
		var userID int64

		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			s.logger.Error(err)
			return
		}
		defer func() {
			_ = r.Body.Close()
		}()

		orderUid, err = strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !utils.Valid(orderUid) {
			http.Error(w, "invalid order format", http.StatusUnprocessableEntity)
			return
		}

		userID, err = s.getUserID(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err = s.business.CreateOrder(r.Context(), userID, orderUid); err != nil {
			if errors.Is(err, customerrors.ErrCurrUserUploaded) {
				w.WriteHeader(http.StatusOK)
				return
			} else if errors.Is(err, customerrors.ErrAnotherUserUploaded) {
				w.WriteHeader(http.StatusConflict)
				return
			} else {
				s.logger.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusAccepted)
		return
	}
}

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
		s.logger.Debug(domenOrders)

		for _, o := range domenOrders {
			orders = append(orders, models.Order(o))
		}

		s.logger.Debug(orders)
		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		s.writeJSONResp(orders, w)
	}
}

func (s *APIServer) getBalance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (s *APIServer) withdraw() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (s *APIServer) getWithdraws() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (s *APIServer) getOrderPaginateList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

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
