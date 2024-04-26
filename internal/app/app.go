package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go_url_chortener_api/internal/config"
	"go_url_chortener_api/internal/env"
	"go_url_chortener_api/internal/http-server/handlers/auth/signin"
	"go_url_chortener_api/internal/http-server/handlers/auth/signup"
	"go_url_chortener_api/internal/http-server/handlers/del"
	"go_url_chortener_api/internal/http-server/handlers/redirect"
	"go_url_chortener_api/internal/http-server/handlers/refresh"
	"go_url_chortener_api/internal/http-server/handlers/url/save"
	"go_url_chortener_api/internal/http-server/middleware/myJwt"
	srv "go_url_chortener_api/internal/http-server/server"
	"go_url_chortener_api/internal/lib/hash"
	"go_url_chortener_api/internal/lib/logger/sl"
	"go_url_chortener_api/internal/lib/logger/slogpretty"
	"go_url_chortener_api/internal/storage/postgres"
	"log/slog"
	"os"
)

func Run(cfg *config.Config) {

	log := setupLogger(cfg.Env)
	log.Info("Starting program...", slog.String("env", cfg.Env))

	storage, err := postgres.NewStorage(&cfg.Storage)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		return
	}

	hasher := hash.NewSHA1Hasher(env.Salt)

	router := getRouter(log, storage, hasher)

	log.Info("starting server...",
		slog.String("address", cfg.HttpServer.Address+":"+cfg.HttpServer.Port),
	)

	server := srv.NewServer(&cfg.HttpServer, router)

	if err := server.Run(); err != nil {
		log.Error("failed to start sever", sl.Err(err))
		return
	}

}

func getRouter(log *slog.Logger, storage *postgres.Storage, hasher *hash.SHA1Hasher) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/auth", func(r chi.Router) {
		r.Post("/signup", signup.New(log, storage, hasher))
		r.Post("/signin", signin.New(log, storage, hasher))
		r.Get("/refresh", refresh.New(log, storage))
	})

	router.Route("/url", func(r chi.Router) {
		r.Use(myJwt.JwtMiddleware(log))
		r.Post("/", save.New(log, storage))
		r.Delete("/{alias}", del.New(log, storage))
	})
	router.Get("/{alias}", redirect.New(log, storage))
	return router
}

func setupLogger(environment string) *slog.Logger {
	var log *slog.Logger
	switch environment {
	case env.EnvLocal:
		log = setupPrettySlog()
	case env.EnvDev:
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
