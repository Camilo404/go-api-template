// Package server wraps http.Server with graceful shutdown driven by
// SIGINT/SIGTERM. In-flight requests are given ShutdownWait to finish
// before the process exits, so deploys do not drop connections.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Camilo404/go-api-template/internal/config"
)

// Server bundles the configured http.Server with the dependencies it
// needs at runtime.
type Server struct {
	cfg    *config.Config
	logger *slog.Logger
	h      http.Handler
}

// New constructs a Server. The handler is typically the result of
// router.New(...).
func New(cfg *config.Config, logger *slog.Logger, h http.Handler) *Server {
	return &Server{cfg: cfg, logger: logger, h: h}
}

// Run blocks until the server stops. It returns nil on a clean
// shutdown and the underlying error otherwise.
func (s *Server) Run(_ context.Context) error {
	srv := &http.Server{
		Addr:         ":" + s.cfg.Port,
		Handler:      s.h,
		ReadTimeout:  s.cfg.ReadTimeout,
		WriteTimeout: s.cfg.WriteTimeout,
		IdleTimeout:  s.cfg.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("server_starting",
			slog.String("addr", srv.Addr),
			slog.String("env", s.cfg.Env),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-stop:
		s.logger.Info("shutdown_signal_received", slog.String("signal", sig.String()))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownWait)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("shutdown_failed", slog.Any("error", err))
		return err
	}
	// Drain the goroutine result so we don't leak it.
	<-errCh
	s.logger.Info("server_stopped")
	return nil
}

// EnsureShutdownTimeout returns the configured wait so callers can
// reuse it in tests if needed.
func (s *Server) ShutdownTimeout() time.Duration { return s.cfg.ShutdownWait }
