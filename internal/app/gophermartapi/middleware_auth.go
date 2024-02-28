package gophermartapi

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ctxUserID struct{}

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
		span := trace.SpanFromContext(r.Context())
		span.SetAttributes(
			attribute.Key("userID").String(strconv.FormatInt(userID, 10)),
		)

		ctx := context.WithValue(r.Context(), ctxUserID{}, userID)
		s.logger.Debugf("Authorize USER.ID=%v", userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *APIServer) getUserID(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value(ctxUserID{}).(int64)
	if !ok {
		s.logger.Warning("user id not found in context")
		return 0, fmt.Errorf("user id not found in context")
	}
	return userID, nil
}
