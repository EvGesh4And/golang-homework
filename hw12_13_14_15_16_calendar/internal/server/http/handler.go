package internalhttp

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
)

func (s *Server) event(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("обработка запроса события", "method", r.Method, "URL", r.URL.Path)

	switch r.Method {
	case http.MethodPost:
		s.logger.Debug("обработка POST запроса на создание события")
		s.handleCreateEvent(w, r)
	case http.MethodPut:
		s.logger.Debug("обработка PUT запроса на обновление события")
		s.handleUpdateEvent(w, r)
	case http.MethodDelete:
		s.logger.Debug("обработка DELETE запроса на удаление события")
		s.handleDeleteEvent(w, r)
	case http.MethodGet:
		s.logger.Debug("обработка GET запроса на получение событий")
		s.handleGetEvents(w, r)
	default:
		s.logger.Error("метод не поддерживается", "HTTPmethod", r.Method)
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	if !s.checkContentType(w, r) {
		return
	}

	event, err := s.getEventFromBody(r)
	if err != nil {
		http.Error(w, server.ErrInvalidEventData.Error(), http.StatusBadRequest)
		return
	}

	s.logger.Debug("попытка создать событие", "method", "handleCreateEvent",
		"eventID", event.ID, "userID", event.UserID)

	if err := s.app.CreateEvent(r.Context(), event); err != nil {
		s.logger.Error("ошибка при создании события", "method", "handleCreateEvent",
			"eventID", event.ID, "userID", event.UserID, "error", err)
		http.Error(w, server.ErrCreateEvent.Error(), http.StatusInternalServerError)
		return
	}

	s.logger.Info("событие успешно создано", "method", "handleCreateEvent",
		"eventID", event.ID, "userID", event.UserID)
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "created", "eventID": event.ID.String()})
}

func (s *Server) handleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	if !s.checkContentType(w, r) {
		return
	}

	uuID, err := s.getEventIDFromBody(r, w)
	if err != nil {
		return
	}

	event, err := s.getEventFromBody(r)
	if err != nil {
		http.Error(w, server.ErrInvalidEventData.Error(), http.StatusBadRequest)
		return
	}

	event.ID = uuID
	s.logger.Debug("попытка обновления события", "method", "handleUpdateEvent",
		"eventID", event.ID, "userID", event.UserID)

	if err := s.app.UpdateEvent(r.Context(), uuID, event); err != nil {
		s.logger.Error("ошибка при обновлении события", "method", "handleUpdateEvent",
			"eventID", event.ID, "userID", event.UserID, "error", err)
		http.Error(w, server.ErrUpdateEvent.Error(), http.StatusInternalServerError)
		return
	}

	s.logger.Info("событие успешно обновлено", "method", "handleUpdateEvent",
		"eventID", event.ID, "userID", event.UserID)
	w.WriteHeader(http.StatusNoContent)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "updated",
		"eventID": event.ID.String(),
	})
}

func (s *Server) handleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	uuID, err := s.getEventIDFromBody(r, w)
	if err != nil {
		return
	}

	s.logger.Debug("попытка удалить событие", "method", "handleDeleteEvent", "eventID", uuID)

	if err := s.app.DeleteEvent(r.Context(), uuID); err != nil {
		s.logger.Error("ошибка при удалении события", "method", "handleDeleteEvent",
			"eventID", uuID.String(), "error", err)
		http.Error(w, server.ErrDeleteEvent.Error(), http.StatusInternalServerError)
		return
	}

	s.logger.Info("событие успешно удалено", "method", "handleDeleteEvent",
		"eventID", uuID.String())
	w.WriteHeader(http.StatusNoContent)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "deleted",
		"eventID": uuID.String(),
	})
}

func (s *Server) handleGetEvents(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		s.logger.Error("неверный формат времени начала", "method", "handleGetEvents",
			"error", err)
		http.Error(w, server.ErrInvalidStartPeriod.Error(), http.StatusBadRequest)
		return
	}

	var events []storage.Event
	period := r.URL.Query().Get("period")

	s.logger.Debug("попытка получения событий", "method", "GetEvents",
		"start", start.Format(time.RFC3339), "period", period)

	switch period {
	case "day":
		events, err = s.app.GetEventsDay(r.Context(), start)
	case "week":
		events, err = s.app.GetEventsWeek(r.Context(), start)
	case "month":
		events, err = s.app.GetEventsMonth(r.Context(), start)
	default:
		s.logger.Error("неверный период", "method", "handleGetEvents", "period", period)
		http.Error(w, server.ErrInvalidPeriod.Error(), http.StatusBadRequest)
		return
	}

	if err != nil {
		s.logger.Error("ошибка при получении событий", "method", "handleGetEvents",
			"start", start.Format(time.RFC3339), "period", period, "error", err)
		http.Error(w, server.ErrEventRetrieval.Error(), http.StatusInternalServerError)
		return
	}

	s.logger.Info("успешно получены события", "method", "handleGetEvents", "count",
		len(events), "period", period, "start", start.Format(time.RFC3339))
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(events)
}

func (s *Server) getEventFromBody(r *http.Request) (storage.Event, error) {
	var event storage.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		s.logger.Error("не удалось распарсить тело запроса в event", "error", err)
		return storage.Event{}, err
	}
	s.logger.Debug("успешно распарсено тело запроса в event",
		"eventID", event.ID, "userID", event.UserID)
	return event, nil
}

func (s *Server) getEventIDFromBody(r *http.Request, w http.ResponseWriter) (uuid.UUID, error) {
	s.logger.Debug("попытка извлечь ID события из параметров запроса", "method", "getEventIDFromBody")
	id := r.URL.Query().Get("id")
	if id == "" {
		s.logger.Error(server.ErrMissingEventID.Error(), "id", nil)
		http.Error(w, server.ErrMissingEventID.Error(), http.StatusBadRequest)
		return uuid.Nil, server.ErrMissingEventID
	}
	uuID, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error(server.ErrInvalidEventID.Error(), "id", id, "error", err)
		http.Error(w, server.ErrInvalidEventID.Error(), http.StatusBadRequest)
		return uuid.Nil, server.ErrInvalidEventID
	}
	s.logger.Debug("успешно извлечён ID из параметров запроса", "eventID", uuID.String())
	return uuID, nil
}

func (s *Server) checkContentType(w http.ResponseWriter, r *http.Request) bool {
	const requiredContentType = "application/json"
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, requiredContentType) {
		s.logger.Error(server.ErrInvalidContentType.Error(), "receivedContentType", contentType)
		http.Error(w, server.ErrInvalidContentType.Error(), http.StatusBadRequest)
		return false
	}
	s.logger.Debug("валидный Content-Type", "Content-Type", contentType)
	return true
}
