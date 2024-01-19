package gophermartapi

import (
	"context"
	"net/http"
)

type CtxKey struct{}

const (
	authHeader = "Authorization"
)

func (s *APIServer) authMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authH := r.Header.Get(authHeader)
		if authH == "" {
			s.logger.Debugln("Auth header not set")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, err := s.auth.ParseToken(authH)

		if err != nil {
			s.logger.Debugf("ParseToken failed: %s", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), CtxKey{}, userID)
		s.logger.Debugf("Authorize USER.ID=%v", userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
