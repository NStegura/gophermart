package gophermartapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

type APIServer struct {
	address     string
	respTimeout time.Duration
	business    Business
	auth        Auth

	router *chi.Mux

	logger *logrus.Logger
}

func New(address string, business Business, auth Auth, logger *logrus.Logger) *APIServer {
	return &APIServer{
		address:     address,
		respTimeout: time.Minute,
		business:    business,
		auth:        auth,
		router:      chi.NewRouter(),
		logger:      logger,
	}
}

func (s *APIServer) Start() error {
	s.configRouter()

	s.logger.Infof("starting APIServer %s", s.address)
	if err := http.ListenAndServe(s.address, s.router); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func (s *APIServer) configRouter() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	s.router.Use(middleware.Timeout(s.respTimeout))

	s.router.Get(`/ping`, s.ping())

	s.router.Group(s.baseRouter)
	s.router.Route(`/api/user`, func(r chi.Router) {
		r.Group(s.authRouter)
		r.Group(s.apiRouter)
	})
}

func (s *APIServer) authRouter(r chi.Router) {
	r.Post(`/register`, s.register())
	r.Post(`/login`, s.login())
}

func (s *APIServer) apiRouter(r chi.Router) {
	r.Use(s.authMiddleware)
	r.Post(`/orders`, s.createOrder())
	r.Get(`/check_auth`, s.ping())
}

func (s *APIServer) baseRouter(r chi.Router) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("root."))
	})
	r.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("test")
	})
	r.Get(`/ping`, s.ping())
}
