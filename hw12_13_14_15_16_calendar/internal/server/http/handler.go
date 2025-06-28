package internalhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /event", s.checkContentTypeMiddleware(http.HandlerFunc(s.CreateEvent)))
	mux.Handle("PUT /event", s.checkContentTypeMiddleware(http.HandlerFunc(s.UpdateEvent)))
	mux.Handle("DELETE /event", http.HandlerFunc(s.DeleteEvent))
	mux.Handle("GET /event/day", http.HandlerFunc(s.GetEventsDay))
	mux.Handle("GET /event/week", http.HandlerFunc(s.GetEventsWeek))
	mux.Handle("GET /event/month", http.HandlerFunc(s.GetEventsMonth))

	return mux
}

// CreateEvent handles event creation request.
func (s *Server) CreateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := logger.WithLogMethod(r.Context(), "CreateEvent")

	event, err := s.getEventFromBody(ctx, r)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		http.Error(w, server.ErrInvalidEventData.Error(), http.StatusBadRequest)
		return
	}

	ctx = logger.WithLogEventID(ctx, event.ID)

	s.logger.DebugContext(ctx, "попытка создать событие")

	if err := s.app.CreateEvent(ctx, event); err != nil {
		s.checkError(w, err, server.ErrCreateEvent)
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return
	}

	s.logger.InfoContext(ctx, "событие успешно создано")
	w.WriteHeader(http.StatusCreated)
}

// UpdateEvent handles event update request.
func (s *Server) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := logger.WithLogMethod(r.Context(), "UpdateEvent")

	uuID, err := s.getEventIDFromBody(ctx, r)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx = logger.WithLogEventID(ctx, uuID)

	event, err := s.getEventFromBody(ctx, r)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		http.Error(w, server.ErrInvalidEventData.Error(), http.StatusBadRequest)
		return
	}
	// Поля ID у события может быть пустым
	event.ID = uuID

	s.logger.DebugContext(ctx, "попытка обновления события")

	if err := s.app.UpdateEvent(ctx, uuID, event); err != nil {
		s.checkError(w, err, server.ErrUpdateEvent)
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return
	}

	s.logger.InfoContext(ctx, "событие успешно обновлено")
	w.WriteHeader(http.StatusNoContent)
}

// DeleteEvent handles event deletion request.
func (s *Server) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	ctx := logger.WithLogMethod(r.Context(), "DeleteEvent")

	uuID, err := s.getEventIDFromBody(ctx, r)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		http.Error(w, server.ErrInvalidEventData.Error(), http.StatusBadRequest)
		return
	}

	ctx = logger.WithLogEventID(ctx, uuID)

	s.logger.DebugContext(ctx, "попытка удалить событие")

	if err := s.app.DeleteEvent(ctx, uuID); err != nil {
		s.checkError(w, err, server.ErrDeleteEvent)
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		return
	}

	s.logger.InfoContext(ctx, "событие успешно удалено")
	w.WriteHeader(http.StatusNoContent)
}

// GetEventsDay returns events for a single day.
func (s *Server) GetEventsDay(w http.ResponseWriter, r *http.Request) {
	s.handleGetEvents(w, r, "Day", s.app.GetEventsDay)
}

// GetEventsWeek returns events for a week.
func (s *Server) GetEventsWeek(w http.ResponseWriter, r *http.Request) {
	s.handleGetEvents(w, r, "Week", s.app.GetEventsWeek)
}

// GetEventsMonth returns events for a month.
func (s *Server) GetEventsMonth(w http.ResponseWriter, r *http.Request) {
	s.handleGetEvents(w, r, "Month", s.app.GetEventsMonth)
}

func (s *Server) handleGetEvents(
	w http.ResponseWriter,
	r *http.Request,
	period string,
	getEventsFunc func(ctx context.Context, start time.Time) ([]storage.Event, error),
) {
	ctx := logger.WithLogMethod(r.Context(), "GetEvents"+period)

	startStr := r.URL.Query().Get("start")
	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		http.Error(w, server.ErrInvalidStartPeriod.Error(), http.StatusBadRequest)
		return
	}

	s.logger.DebugContext(ctx, "попытка получить события")

	events, err := getEventsFunc(r.Context(), start)
	if err != nil {
		s.logger.ErrorContext(logger.ErrorCtx(ctx, err), err.Error())
		http.Error(w, server.ErrEventRetrieval.Error(), http.StatusInternalServerError)
		return
	}

	eventsDTO := make([]storage.EventDTO, len(events))
	for i := range events {
		eventsDTO[i] = storage.ToDTO(events[i])
	}

	s.logger.InfoContext(ctx, "успешно получены события", "count", len(events))
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(eventsDTO)
}
