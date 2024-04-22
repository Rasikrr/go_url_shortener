package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go_url_chortener_api/internal/config"
	"go_url_chortener_api/internal/http-server/handlers/del"
	"go_url_chortener_api/internal/http-server/handlers/redirect"
	"go_url_chortener_api/internal/http-server/handlers/url/save"
	"go_url_chortener_api/internal/lib/logger/sl"
	"go_url_chortener_api/internal/lib/logger/slogpretty"
	"go_url_chortener_api/internal/storage/postgres"
	"log/slog"
	"net/http"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log.Info("Starting program...", slog.String("env", cfg.Env))

	storage, err := postgres.NewStorage(&cfg.Storage)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		return
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url", save.New(log, storage))
	router.Get("/{alias}", redirect.New(log, storage))
	router.Delete("/{alias}", del.New(log, storage))

	log.Info("starting server...",
		slog.String("address", cfg.HttpServer.Address+":"+cfg.HttpServer.Port),
	)

	srv := &http.Server{
		Addr:         cfg.HttpServer.Address + ":" + cfg.HttpServer.Port,
		Handler:      router,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		ReadTimeout:  cfg.HttpServer.Timeout,
	}
	srv.ListenAndServe()

	log.Error("server stopped")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
