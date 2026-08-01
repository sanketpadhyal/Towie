package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sanketpadhyal/towie/internal/config"
)

type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
}

func New(cfg *config.Config, handler http.Handler) *Server {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		},
		shutdownTimeout: cfg.Server.ShutdownTimeout,
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
	defer cancel()
	return s.httpServer.Shutdown(shutdownCtx)
}

func (s *Server) Addr() string {
	return s.httpServer.Addr
}
