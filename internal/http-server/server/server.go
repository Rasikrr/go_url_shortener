package server

import (
	"go_url_chortener_api/internal/config"
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg *config.HttpServer, handler http.Handler) *Server {
	httpServer := &http.Server{
		Addr:         cfg.Address + ":" + cfg.Port,
		Handler:      handler,
		IdleTimeout:  cfg.IdleTimeout,
		WriteTimeout: cfg.Timeout,
		ReadTimeout:  cfg.Timeout,
	}
	srv := &Server{
		httpServer: httpServer,
	}
	return srv
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}
