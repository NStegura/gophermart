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

	"github.com/NStegura/gophermart/internal/monitoring/logger"
	"github.com/NStegura/gophermart/internal/monitoring/tracer"

	"github.com/NStegura/gophermart/internal/clients/accrual"
	"github.com/NStegura/gophermart/internal/services/jobs/accrualsync"

	"github.com/NStegura/gophermart/internal/app/gophermartapi"
	"github.com/NStegura/gophermart/internal/repo"
	"github.com/NStegura/gophermart/internal/services/auth"
	"github.com/NStegura/gophermart/internal/services/business"
)

const (
	timeoutShutdown = time.Second * 10
	rateLimit       = 5
	frequency       = time.Duration(15) * time.Second
	serviceName     = "Gophermart"
)

func runApp() error {
	ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	config := gophermartapi.NewConfig()
	err := config.ParseFlags()
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	logg, err := logger.Init(config.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}

	_, err = tracer.Init(config.TracerURL, serviceName)
	if err != nil {
		return fmt.Errorf("failed to init tracer: %w", err)
	}

	db, err := repo.New(
		ctx,
		config.DatabaseDSN,
		logg,
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
		defer logg.Info("closed DB")
		defer wg.Done()
		<-ctx.Done()

		db.Shutdown(ctx)
	}()

	server := gophermartapi.New(
		config.RunAddress,
		business.New(db, logg),
		auth.New(config.SecretKey, logg),
		logg,
	)

	accrualCli, err := accrual.New(config.AccrualAddr, logg)
	if err != nil {
		return fmt.Errorf("failed to init accrualCli: %w", err)
	}

	accrualJob := accrualsync.New(
		frequency,
		rateLimit,
		db,
		accrualCli,
		logg,
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

	go func(errs chan<- error) {
		if err = accrualJob.Start(ctx); err != nil {
			errs <- fmt.Errorf("accrualJob has failed: %w", err)
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
	if err := runApp(); err != nil {
		log.Fatal(err)
	}
}
