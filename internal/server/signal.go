package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Run(s *Server, log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sighupCh := make(chan os.Signal, 1)
	signal.Notify(sighupCh, syscall.SIGHUP)
	defer signal.Stop(sighupCh)

	errCh := make(chan error, 1)
	go func() {
		log.Info("server started", "addr", s.Addr())
		if err := s.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	for {
		select {
		case err := <-errCh:
			return err
		case <-sighupCh:
			log.Info("configuration reload signal (SIGHUP) received")
		case <-ctx.Done():
			stop()
			log.Info("shutdown signal received, draining connections")
			if err := s.Shutdown(context.Background()); err != nil {
				log.Error("shutdown error", "error", err)
				return err
			}
			log.Info("server stopped")
			return nil
		}
	}
}
