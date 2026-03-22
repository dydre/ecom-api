package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecom-api/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type app struct {
	config *config.Config
	logger *slog.Logger
}

func (a *app) mount(handlerTimeout time.Duration) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(handlerTimeout))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok\n"))
	})

	return r
}

func (a *app) run(h http.Handler) error {
	server := &http.Server{
		Addr:              ":" + a.config.HTTP.Port,
		Handler:           h,
		WriteTimeout:      a.config.HTTP.WriteTimeout,
		ReadTimeout:       a.config.HTTP.ReadTimeout,
		ReadHeaderTimeout: a.config.HTTP.ReadHeaderTimeout,
		IdleTimeout:       a.config.HTTP.IdleTimeout,
	}

	go func() {
		a.logger.Info("🚀 server started", slog.String("addr", ":"+a.config.HTTP.Port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("❌ server failed to start", slog.Any("error", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("⏳ shutdown signal received, stopping server...")

	ctx, cancel := context.WithTimeout(context.Background(), a.config.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return err
	}

	a.logger.Info("✅ server stopped gracefully")
	return nil
}

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("🔧 config loaded", slog.String("env", cfg.Env))

	api := &app{
		config: cfg,
		logger: log,
	}

	handlerTimeout := cfg.HTTP.WriteTimeout - 5*time.Second
	h := api.mount(handlerTimeout)

	if err := api.run(h); err != nil {
		log.Error("💥 server shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	slog.SetDefault(log)
	return log
}
