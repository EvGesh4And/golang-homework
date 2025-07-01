package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
)

func (s *Server) getEventFromBody(ctx context.Context, r *http.Request) (storage.Event, error) {
	var event storage.EventDTO
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		return storage.Event{}, logger.AddPrefix(ctx, err)
	}
	s.logger.DebugContext(ctx, "request body successfully parsed into event")
	return storage.FromDTO(event), nil
}

func (s *Server) getEventIDFromBody(ctx context.Context, r *http.Request) (uuid.UUID, error) {
	ctx = s.setLogCompMeth(ctx, "getEventIDFromBody")
	s.logger.DebugContext(ctx, "attempting to extract event ID from request parameters")
	id := r.URL.Query().Get("id")
	if id == "" {
		return uuid.Nil, logger.AddPrefix(ctx, server.ErrMissingEventID)
	}
	uuID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, logger.AddPrefix(ctx, server.ErrInvalidEventID)
	}
	ctx = logger.WithLogEventID(ctx, uuID)
	s.logger.DebugContext(ctx, "event ID successfully extracted from request parameters")
	return uuID, nil
}

func (s *Server) checkError(w http.ResponseWriter, err error, internalServerError error) {
	var ve *storage.ErrInvalidEvent
	if errors.As(err, &ve) {
		http.Error(w, ve.Error(), http.StatusBadRequest)
		return
	}

	if errors.Is(err, storage.ErrIDRepeated) {
		http.Error(w, storage.ErrIDRepeated.Error(), http.StatusConflict)
		return
	}

	if errors.Is(err, storage.ErrIDNotExist) {
		http.Error(w, storage.ErrIDNotExist.Error(), http.StatusNotFound)
		return
	}

	if errors.Is(err, storage.ErrDateBusy) {
		http.Error(w, storage.ErrDateBusy.Error(), http.StatusConflict)
		return
	}

	http.Error(w, internalServerError.Error(), http.StatusInternalServerError)
}
