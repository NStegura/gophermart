package gophermartapi

import (
	"encoding/json"
	"fmt"

	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/NStegura/gophermart/docs"
)

const (
	contType string = "Content-Type"
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

// Start godoc
//
//	@title						Gophermart API
//	@version					1.0
//	@description				This is a Gophermart server.
//	@BasePath					/
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
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
	r.Get(`/orders`, s.getOrderList())
	r.Get(`/orders/paginate`, s.getOrderPaginateList())
	r.Get(`/balance`, s.getBalance())
	r.Post(`/balance/withdraw`, s.createWithdraw())
	r.Get(`/withdrawals`, s.getWithdrawals())
}

func (s *APIServer) baseRouter(r chi.Router) {
	r.Mount("/swagger", httpSwagger.WrapHandler)
	r.Get(`/ping`, s.ping())
}

func (s *APIServer) writeJSONResp(resp any, w http.ResponseWriter) {
	w.Header().Set(contType, "application/json")

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(jsonResp); err != nil {
		s.logger.Error(err)
	}
}
