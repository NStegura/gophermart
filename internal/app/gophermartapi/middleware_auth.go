package gophermartapi

import (
	"context"
	"fmt"
	"net/http"
)

type CtxUserID struct{}

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
		ctx := context.WithValue(r.Context(), CtxUserID{}, userID)
		s.logger.Debugf("Authorize USER.ID=%v", userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *APIServer) getUserID(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value(CtxUserID{}).(int64)
	if !ok {
		s.logger.Warning("user id not found in context")
		return 0, fmt.Errorf("user id not found in context")
	}
	return userID, nil
}
