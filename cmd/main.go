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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	httpPort = "8080"

	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = time.Minute
)

type config struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	dsn string
}

type app struct {
	config config
	logger *slog.Logger
}

func (a *app) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	return r
}

func (a *app) run(h http.Handler) error {
	server := &http.Server{
		Addr:              a.config.addr,
		Handler:           h,
		WriteTimeout:      writeTimeout,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
	}

	go func() {
		a.logger.Info("🚀 HTTP-сервер запущен", slog.String("addr", a.config.addr))

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("❌ Ошибка запуска сервера", slog.Any("error", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("🛑 Получен сигнал завершения, останавливаем сервер...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		// Передаём ошибку наверх — main решит, что с ней делать.
		return err
	}

	a.logger.Info("✅ Сервер остановлен")
	return nil
}

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)

	slog.SetDefault(logger)

	cfg := config{
		addr: httpPort,
		db:   dbConfig{},
	}

	api := &app{
		config: cfg,
		logger: logger,
	}

	h := api.mount()

	if err := api.run(h); err != nil {

		logger.Error("❌ Ошибка при остановке сервера", slog.Any("error", err))
		os.Exit(1)
	}
}
