package internalhttp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server"
)

// Server provides HTTP access to the calendar application.
type Server struct {
	logger     *slog.Logger
	app        server.Application
	httpServer *http.Server
	handler    http.Handler
}

// Handler returns http.Handler used by the server.
func (s *Server) Handler() http.Handler {
	return s.handler
}

// NewServerHTTP creates and configures a new HTTP server.
func NewServerHTTP(host string, port int, logger *slog.Logger, app server.Application) *Server {
	s := &Server{
		logger: logger,
		app:    app,
	}

	mux := s.routes()

	wrapped := s.loggingMiddleware(mux)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", host, port),
		Handler:           wrapped,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	s.httpServer = httpServer
	s.handler = wrapped
	return s
}

// Start runs the HTTP server.
func (s *Server) Start() error {
	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server run error: %w", err)
	}
	return nil
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
