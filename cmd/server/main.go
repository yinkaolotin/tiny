package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yinkaolotin/tiny/internal/config"
	"github.com/yinkaolotin/tiny/internal/httpapi"
	"github.com/yinkaolotin/tiny/internal/logger"
	"github.com/yinkaolotin/tiny/internal/storage"
	"github.com/yinkaolotin/tiny/internal/worker"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	store := storage.NewMemoryStore()
	handler := httpapi.New(store, log)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/ready", handler.Ready)
	mux.HandleFunc("/items", handler.Items)

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	worker.StartCleanup(ctx, store, log)

	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(shutdownCtx)
	log.Info().Msg("server stopped gracefully")
}
