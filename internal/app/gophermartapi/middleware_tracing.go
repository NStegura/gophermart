package gophermartapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func (s *APIServer) tracingMiddleware(h http.Handler) http.Handler {
	return otelhttp.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)

			span := trace.SpanFromContext(r.Context())
			routePattern := chi.RouteContext(r.Context()).RoutePattern()
			span.SetName(routePattern)
			labeler, ok := otelhttp.LabelerFromContext(r.Context())
			if ok {
				labeler.Add(attribute.Key("route").String(routePattern))
			}
		}),
		"",
	)
}
