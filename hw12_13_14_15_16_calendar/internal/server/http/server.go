package internalhttp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
)

type Server struct { // TODO
	logger     Logger
	app        Application
	httpServer *http.Server
	ErrCh      chan error
}

type Logger interface {
	Error(module string, msg string, args ...any)
	Warn(module string, msg string, args ...any)
	Info(module string, msg string, args ...any)
	Debug(module string, msg string, args ...any)
	Printf(msg string, args ...any)
}

type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, id uuid.UUID, event storage.Event) error
	DeleteEvent(ctx context.Context, id uuid.UUID) error
	GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error)
	GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error)
	GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error)
}

func NewServerHTTP(host string, port int, logger Logger, app Application) *Server {
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

	s.ErrCh = make(chan error, 1)
	return s
}

func (s *Server) Start() error {
	// Запускаем ListenAndServe в текущем потоке — main его обернёт в go func
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
