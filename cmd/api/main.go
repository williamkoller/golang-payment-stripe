package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/williamkoller/golang-payment-stripe/internal/app/saga"
	"github.com/williamkoller/golang-payment-stripe/internal/app/service"
	"github.com/williamkoller/golang-payment-stripe/internal/infra/config"
	"github.com/williamkoller/golang-payment-stripe/internal/infra/http/router"
	"github.com/williamkoller/golang-payment-stripe/internal/infra/logger"
	"github.com/williamkoller/golang-payment-stripe/internal/infra/repo/memory"
	stripeinfra "github.com/williamkoller/golang-payment-stripe/internal/infra/stripe"
)

func main() {
	cfg := config.Load()
	zl := logger.New(cfg)

	repo := memory.NewPaymentRepo()
	stripeClient := stripeinfra.NewClient(cfg, zl)

	paymentSaga := saga.NewPaymentSaga(zl, repo, stripeClient, cfg)
	paymentSvc := service.NewPaymentService(zl, repo, paymentSaga)

	engine := router.Build(zl, cfg, paymentSvc, stripeClient, repo)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		zl.Sugar().Infow("server_start", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zl.Sugar().Fatalw("listen", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	zl.Sugar().Info("server_stopped")
}
