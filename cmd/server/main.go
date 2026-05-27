package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yinkaolotin/tiny/internal/config"
	"github.com/yinkaolotin/tiny/internal/httpapi"
	"github.com/yinkaolotin/tiny/internal/logger"
	"github.com/yinkaolotin/tiny/internal/metrics"
	"github.com/yinkaolotin/tiny/internal/storage"
	"github.com/yinkaolotin/tiny/internal/worker"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err := storage.RunMigrations(cfg); err != nil {
			log.Fatal().Err(err).Msg("migration failed")
		}
		log.Info().Msg("migration completed")
		return
	}

	var store storage.Store

	switch cfg.StorageBackend {
	case "postgres":
		store, err = storage.NewPostgresStore(cfg.DatabaseURL())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to init postgres store")
		}
	case "file":
		store, err = storage.NewFileStore(cfg.DataDir)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to init filestore")
		}
	case "memory":
		store = storage.NewMemoryStore()
	default:
		log.Fatal().Str("storage_backend", cfg.StorageBackend).Msg("unknown storage backend")
	}

	handler := httpapi.New(store, log)
	metrics.Register()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/ready", handler.Ready)
	mux.HandleFunc("/items", handler.Items)
	mux.Handle("/metrics", promhttp.Handler())

	root := httpapi.MetricsMiddleware(mux)

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: root,
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
