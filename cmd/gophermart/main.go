package main

import (
	"context"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/api"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	// init app and start server
	serverAddress, err := gophermart.Setup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{
		Addr:    *serverAddress,
		Handler: gophermart.Routes(),
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Log.Infof("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal(err)
		}
	}()

	// check accrual
	wg.Add(1)
	go gophermart.CheckAccrual(ctx, &wg)

	// gracefully shutdown by signal
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-c
		cancel()
		// shutdown server
		if err := srv.Shutdown(ctx); err != nil {
			logger.Log.Fatalf("Server shutdown failed: %v", err)
		}
		if err := api.Repo.Close(); err != nil {
			logger.Log.Fatalf("Channel Jobs shutdown failed: %v", err)
		}
	}()

	wg.Wait()
	logger.Log.Infoln("Successful shutdown")
}
