package main

import (
	"context"
	"net/http"
	"service/internal/initenv"
	"service/internal/middleware"
	"service/internal/wallet"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xlab/closer"
)

func main() {

	ctx := context.Background()
	lg, cfgAdr, err := initenv.Start(ctx)
	if err != nil {
		lg.FatalCtx(ctx, "Error starting initenv", err)
	}
	defer closer.Close()

	router := chi.NewRouter()
	walletHandler := wallet.NewHandler(lg, ctx, cfgAdr)

	router.Use(middleware.ContextRequestMiddleware)
	router.Post("/api/v1/wallet", walletHandler.HandleWalletOperation)
	router.Get("/api/v1/balance/{id}", walletHandler.GetWalletBalance)

	closer.Bind(func() {

		walletHandler.Close()
		time.Sleep(3 * time.Second)

		lg.InfoCtx(ctx, "Database connection closed")
	})

	err = http.ListenAndServe(cfgAdr.APP_ADR, router)
	if err != nil {
		lg.FatalCtx(ctx, "Error starting server", err)
	}
	lg.InfoCtx(ctx, "Server started")

	closer.Hold()
}
