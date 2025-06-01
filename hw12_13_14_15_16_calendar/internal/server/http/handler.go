package internalhttp

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
)

func (s *Server) event(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateEvent(w, r)
	case http.MethodPut:
		s.handleUpdateEvent(w, r)
	case http.MethodDelete:
		s.handleDeleteEvent(w, r)
	case http.MethodGet:
		s.handleGetEvents(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается", nil)
	}
}

func (s *Server) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	if !s.checkContentType(w, r) {
		return
	}
	var event storage.Event
	if err := getEventFromBody(r, &event); err != nil {
		s.writeError(w, http.StatusBadRequest, "Некорректные данные события", err)
		return
	}
	if err := s.app.CreateEvent(r.Context(), event); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Ошибка при создании события ID:"+event.ID.String(), nil)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	if !s.checkContentType(w, r) {
		return
	}
	var event storage.Event
	if err := getEventFromBody(r, &event); err != nil {
		s.writeError(w, http.StatusBadRequest, "Некорректные данные события", err)
		return
	}
	if err := s.app.UpdateEvent(r.Context(), event.ID, event); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Ошибка при обновлении события ID:"+event.ID.String(), nil)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	if !s.checkContentType(w, r) {
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		s.writeError(w, http.StatusBadRequest, "Отсутствует ID события", nil)
		return
	}
	uuID, err := uuid.Parse(id)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Некорректный ID события", err)
		return
	}
	if err := s.app.DeleteEvent(r.Context(), uuID); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Ошибка при удалении события ID:"+uuID.String(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetEvents(w http.ResponseWriter, r *http.Request) {
	if !s.checkContentType(w, r) {
		return
	}
	startStr := r.URL.Query().Get("start")
	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Некорректный формат времени: "+startStr, err)
		return
	}
	period := r.URL.Query().Get("period")
	if period == "" {
		s.writeError(w, http.StatusBadRequest, "Отсутствует период", nil)
		return
	}

	var events []storage.Event
	switch period {
	case "day":
		if events, err = s.app.GetEventsDay(r.Context(), start); err != nil {
			s.writeError(w, http.StatusInternalServerError, "Ошибка при получении событий", nil)
			return
		}
	case "week":
		if events, err = s.app.GetEventsWeek(r.Context(), start); err != nil {
			s.writeError(w, http.StatusInternalServerError, "Ошибка при получении событий", nil)
			return
		}
	case "month":
		if events, err = s.app.GetEventsMonth(r.Context(), start); err != nil {
			s.writeError(w, http.StatusInternalServerError, "Ошибка при получении событий", nil)
			return
		}
	default:
		s.writeError(w, http.StatusBadRequest, "Некорректный период", nil)
	}
	if err := json.NewEncoder(w).Encode(events); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Ошибка кодирования ответа", err)
	}
}

func (s *Server) checkContentType(w http.ResponseWriter, r *http.Request) bool {
	currType := "application/json"
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, currType) {
		s.writeError(w, http.StatusBadRequest, "Content-Type должен быть "+currType, nil)
		return false
	}
	return true
}

func (s *Server) writeError(w http.ResponseWriter, code int, message string, err error) {
	if err != nil {
		switch code {
		case http.StatusBadRequest:
			s.logger.Warn("HTTP", message, err)
		default:
			s.logger.Error("HTTP", message, err)
		}
	} else {
		s.logger.Warn("HTTP", message)
	}
	http.Error(w, message, code)
}

func getEventFromBody(r *http.Request, event *storage.Event) error {
	if err := json.NewDecoder(r.Body).Decode(event); err != nil {
		return err
	}
	return nil
}
