package gophermartapi

import (
	"encoding/json"
	"errors"
	"github.com/NStegura/gophermart/internal/app/gophermartapi/models"
	"github.com/NStegura/gophermart/internal/customerrors"
	"net/http"
)

func (s *APIServer) register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var inputUser models.User

		if err := json.NewDecoder(r.Body).Decode(&inputUser); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		salt := s.auth.GenerateUserSalt(30)
		newPass := s.auth.GeneratePasswordHash(inputUser.Password, salt)

		uID, err := s.business.CreateUser(r.Context(), inputUser.Login, newPass, salt)
		if err != nil {
			if errors.Is(err, customerrors.ErrAlreadyExists) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			s.logger.Debugln(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		token, err := s.auth.GenerateToken(int64(uID))
		w.Header().Set("Authorization", token)
		w.WriteHeader(http.StatusOK)
		return
	}
}

func (s *APIServer) login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var inputUser models.User

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
		token, err := s.auth.GenerateToken(dbUser.ID)
		w.Header().Set("Authorization", token)
		w.WriteHeader(http.StatusOK)
		return
	}
}

func (s *APIServer) createOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (s *APIServer) getOrderList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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
