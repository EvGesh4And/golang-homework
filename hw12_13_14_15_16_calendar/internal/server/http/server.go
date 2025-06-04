package internalhttp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server"
)

type Server struct { // TODO
	logger     *slog.Logger
	app        server.Application
	httpServer *http.Server
	ErrCh      chan error
	handler    http.Handler
}

func (s *Server) Handler() http.Handler {
	return s.handler
}

func NewServerHTTP(host string, port int, logger *slog.Logger, app server.Application) *Server {
	s := &Server{
		logger: logger,
		app:    app,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/event", s.event)

	wrapped := loggingMiddleware(logger)(mux)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", host, port),
		Handler:           wrapped,
		ReadHeaderTimeout: 5 * time.Second,
	}
	s.httpServer = httpServer
	s.handler = wrapped

	s.ErrCh = make(chan error, 1)
	return s
}

func (s *Server) Start() error {
	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("ошибка в работе сервера: %w", err)
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
