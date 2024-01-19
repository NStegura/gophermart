package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/NStegura/gophermart/internal/app/gophermartapi"
	"github.com/NStegura/gophermart/internal/repo"
	"github.com/NStegura/gophermart/internal/services/auth"
	"github.com/NStegura/gophermart/internal/services/business"
)

const (
	timeoutShutdown = time.Second * 10
)

func configureLogger(config *gophermartapi.Config) (*logrus.Logger, error) {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{FullTimestamp: true}
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	logger.SetLevel(level)
	return logger, nil
}

func runRest() error {
	ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	config := gophermartapi.NewConfig()
	err := config.ParseFlags()
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	logger, err := configureLogger(config)
	if err != nil {
		return err
	}

	db, err := repo.New(
		ctx,
		config.DatabaseURI,
		logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create repo: %w", err)
	}

	if err = db.RunMigrations(); err != nil {
		return fmt.Errorf("failed to migrate db: %w", err)
	}

	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
	}()

	wg.Add(1)
	go func() {
		defer logger.Info("closed DB")
		defer wg.Done()
		<-ctx.Done()

		db.Shutdown(ctx)
	}()

	server := gophermartapi.New(
		config.RunAddress,
		business.New(db, logger),
		auth.New(config.SecretKey, logger),
		logger,
	)

	componentsErrs := make(chan error, 1)
	go func(errs chan<- error) {
		if err = server.Start(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			errs <- fmt.Errorf("listen and server has failed: %w", err)
		}
	}(componentsErrs)

	select {
	case <-ctx.Done():
	case err := <-componentsErrs:
		log.Print(err)
		cancelCtx()
	}

	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed to gracefully shutdown the service")
	}()

	return nil
}

func main() {
	if err := runRest(); err != nil {
		log.Fatal(err)
	}
}
