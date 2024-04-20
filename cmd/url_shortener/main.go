package main

import (
	"fmt"
	"go_url_chortener_api/internal/config"
	"go_url_chortener_api/internal/storage/postgres"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
)

func main() {
	cfg := config.MustLoad()
	logger := setupLogger(cfg.Env)
	logger.Info("Starting program...", slog.String("env", cfg.Env))
	db, err := postgres.NewStorage(&cfg.Storage)
	if err != nil {
		logger.Error(err.Error())
		fmt.Println(db)
	}

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
