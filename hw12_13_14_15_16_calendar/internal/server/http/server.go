package internalhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
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
	UpdateEvent(ctx context.Context, id string, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error)
	GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error)
	GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error)
}

func NewServer(host string, port int, logger Logger, app Application) *Server {
	s := &Server{
		logger: logger,
		app:    app,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", s.Hello)
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

func (s *Server) Hello(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	if name == "" {
		name = "world"
	}

	msg := fmt.Sprintf("Hello, %s!", name)
	w.Write([]byte(msg))
}

func (s *Server) event(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// Обработка создания события
		err := s.app.CreateEvent(context.Background(), storage.Event{
			ID:    "550e8400-e29b-41d4-a716-446655440000",
			Title: "sdfsd", Start: time.Now(), End: time.Now().Add(time.Hour), Description: "sdfsd",
			UserID: "550e8400-e29b-41d4-a716-446655440000", TimeBefore: time.Hour,
		})
		if err != nil {
			http.Error(w, "CreateEvent", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	case http.MethodPut:
		// Обработка обновления события
		err := s.app.UpdateEvent(context.Background(), "550e8400-e29b-41d4-a716-446655440000", storage.Event{
			ID:    "550e8400-e29b-41d4-a716-446655440000",
			Title: "sdfsd", Start: time.Now(), End: time.Now().Add(2 * time.Hour), Description: "sdfsd",
			UserID: "550e8400-e29b-41d4-a716-446655440000", TimeBefore: time.Hour,
		})
		if err != nil {
			http.Error(w, "UpdateEvent", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	case http.MethodDelete:
		// Обработка удаления события
		err := s.app.DeleteEvent(context.Background(), "550e8400-e29b-41d4-a716-446655440000")
		if err != nil {
			http.Error(w, "DeleteEvent", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		// Обработка получения событий
		events, err := s.app.GetEventsDay(context.Background(), time.Now())
		if err != nil {
			http.Error(w, "GetEventsDay", http.StatusInternalServerError)
			return
		}

		// Отправка списка событий в ответ
		if err := json.NewEncoder(w).Encode(events); err != nil {
			http.Error(w, "ошибка в json.Encode", http.StatusInternalServerError)
		}
	}
}
