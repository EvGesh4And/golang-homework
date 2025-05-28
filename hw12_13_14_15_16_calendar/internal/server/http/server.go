package internalhttp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/EvGesh4And/hw12_13_14_15_calendar/internal/storage"
)

type Server struct { // TODO
	logger     Logger
	app        Application
	httpServer *http.Server
	ErrCh      chan error
}

type Logger interface {
	Error(msg string, args ...any)
	Warn(msg string, args ...any)
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
}

type Application interface { // TODO
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, id string, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error)
	GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error)
	GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error)
}

func NewServer(host string, port int, logger Logger, app Application) *Server {

	handler := &MyHandler{}
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", handler.Hello)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: mux,
	}

	return &Server{
		logger:     logger,
		app:        app,
		httpServer: httpServer,
		ErrCh:      make(chan error, 1),
	}
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
		return fmt.Errorf("ошибка завершения сервера: %w", err)
	}

	return nil
}

type MyHandler struct{}

func (h *MyHandler) Hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, world!"))
}
