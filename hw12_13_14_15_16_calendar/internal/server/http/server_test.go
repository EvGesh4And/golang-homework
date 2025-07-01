package internalhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	serverpkg "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockApp struct {
	createEvent    func(ctx context.Context, event storage.Event) error
	updateEvent    func(ctx context.Context, id uuid.UUID, event storage.Event) error
	deleteEvent    func(ctx context.Context, id uuid.UUID) error
	getEventsDay   func(ctx context.Context, start time.Time) ([]storage.Event, error)
	getEventsWeek  func(ctx context.Context, start time.Time) ([]storage.Event, error)
	getEventsMonth func(ctx context.Context, start time.Time) ([]storage.Event, error)
}

func (m *mockApp) CreateEvent(ctx context.Context, event storage.Event) error {
	return m.createEvent(ctx, event)
}

func (m *mockApp) UpdateEvent(ctx context.Context, id uuid.UUID, event storage.Event) error {
	return m.updateEvent(ctx, id, event)
}

func (m *mockApp) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	return m.deleteEvent(ctx, id)
}

func (m *mockApp) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return m.getEventsDay(ctx, start)
}

func (m *mockApp) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return m.getEventsWeek(ctx, start)
}

func (m *mockApp) GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return m.getEventsMonth(ctx, start)
}

func TestCreateEvent(t *testing.T) {
	app := &mockApp{
		createEvent: func(ctx context.Context, event storage.Event) error {
			_ = ctx
			_ = event
			return nil
		},
	}

	logger := logger.New("info", os.Stdout, false)
	server := NewServerHTTP("localhost", 8080, logger, app)

	event := storage.Event{
		ID:          uuid.New(),
		Title:       "test event",
		Description: "test desc",
		UserID:      uuid.New(),
		Start:       time.Now().Add(time.Hour),
		End:         time.Now().Add(time.Hour).Add(time.Hour),
		TimeBefore:  10 * time.Minute,
	}

	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/event", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Result().StatusCode)
}

func TestUpdateEvent(t *testing.T) {
	eventID := uuid.New()

	app := &mockApp{
		updateEvent: func(ctx context.Context, id uuid.UUID, event storage.Event) error {
			_ = ctx
			_ = id
			_ = event
			return nil
		},
	}

	logger := logger.New("info", os.Stdout, false)
	server := NewServerHTTP("localhost", 8080, logger, app)

	event := storage.Event{
		Title:       "updated event",
		Description: "desc",
		UserID:      uuid.New(),
		Start:       time.Now().Add(time.Hour),
		End:         time.Now().Add(2 * time.Hour),
		TimeBefore:  5 * time.Minute,
	}

	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPut, "/event?id="+eventID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
}

func TestDeleteEvent(t *testing.T) {
	eventID := uuid.New()

	app := &mockApp{
		deleteEvent: func(ctx context.Context, id uuid.UUID) error {
			_ = ctx
			_ = id
			return nil
		},
	}

	logger := logger.New("info", os.Stdout, false)
	server := NewServerHTTP("localhost", 8080, logger, app)
	req := httptest.NewRequest(http.MethodDelete, "/event?id="+eventID.String(), nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Result().StatusCode)
}

func TestGetEventsDay(t *testing.T) {
	app := &mockApp{
		getEventsDay: func(ctx context.Context, start time.Time) ([]storage.Event, error) {
			_ = ctx
			_ = start
			return []storage.Event{
				{ID: uuid.New(), Title: "Day event"},
			}, nil
		},
	}

	logger := logger.New("info", os.Stdout, false)
	server := NewServerHTTP("localhost", 8080, logger, app)
	req := httptest.NewRequest(http.MethodGet, "/event/day?start=2025-01-01T00:00:00Z", nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestGetEventsWeek(t *testing.T) {
	app := &mockApp{
		getEventsWeek: func(ctx context.Context, start time.Time) ([]storage.Event, error) {
			_ = ctx
			_ = start
			return []storage.Event{
				{ID: uuid.New(), Title: "Week event"},
			}, nil
		},
	}

	logger := logger.New("info", os.Stdout, false)
	server := NewServerHTTP("localhost", 8080, logger, app)
	req := httptest.NewRequest(http.MethodGet, "/event/week?start=2025-01-01T00:00:00Z", nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestGetEventsMonth(t *testing.T) {
	app := &mockApp{
		getEventsMonth: func(ctx context.Context, start time.Time) ([]storage.Event, error) {
			_ = ctx
			_ = start
			return []storage.Event{
				{ID: uuid.New(), Title: "Month event"},
			}, nil
		},
	}

	logger := logger.New("info", os.Stdout, false)
	server := NewServerHTTP("localhost", 8080, logger, app)
	req := httptest.NewRequest(http.MethodGet, "/event/month?start=2025-01-01T00:00:00Z", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestCreateEvent_BadJSON(t *testing.T) {
	app := &mockApp{}
	logger := logger.New("info", os.Stdout, false)
	server := NewServerHTTP("localhost", 8080, logger, app)

	req := httptest.NewRequest(http.MethodPost, "/event", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	assert.Equal(t, serverpkg.ErrInvalidEventData.Error()+"\n", w.Body.String())
}

func TestCreateEvent_StorageError(t *testing.T) {
	app := &mockApp{createEvent: func(ctx context.Context, event storage.Event) error {
		_ = ctx
		_ = event
		return storage.ErrIDRepeated
	}}
	logger := logger.New("info", os.Stdout, false)
	server := NewServerHTTP("localhost", 8080, logger, app)

	event := storage.Event{
		ID:         uuid.New(),
		Title:      "title",
		UserID:     uuid.New(),
		Start:      time.Now().Add(time.Hour),
		End:        time.Now().Add(2 * time.Hour),
		TimeBefore: time.Minute,
	}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/event", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Result().StatusCode)
	assert.Equal(t, storage.ErrIDRepeated.Error()+"\n", w.Body.String())
}

func TestUpdateEvent_InvalidID(t *testing.T) {
	app := &mockApp{}
	logger := logger.New("info", os.Stdout, false)
	server := NewServerHTTP("localhost", 8080, logger, app)

	event := storage.Event{
		Title: "title", UserID: uuid.New(), Start: time.Now().Add(time.Hour),
		End: time.Now().Add(2 * time.Hour),
	}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPut, "/event?id=bad", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	assert.Equal(t, "server.http.getEventIDFromBody: "+serverpkg.ErrInvalidEventID.Error()+"\n", w.Body.String())
}

func TestDeleteEvent_NotFound(t *testing.T) {
	eventID := uuid.New()
	app := &mockApp{deleteEvent: func(ctx context.Context, id uuid.UUID) error {
		_ = ctx
		_ = id
		return storage.ErrIDNotExist
	}}
	logger := logger.New("info", os.Stdout, false)
	server := NewServerHTTP("localhost", 8080, logger, app)

	req := httptest.NewRequest(http.MethodDelete, "/event?id="+eventID.String(), nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)
	assert.Equal(t, storage.ErrIDNotExist.Error()+"\n", w.Body.String())
}

func TestGetEventsDay_InvalidStart(t *testing.T) {
	app := &mockApp{}
	logger := logger.New("info", os.Stdout, false)
	server := NewServerHTTP("localhost", 8080, logger, app)

	req := httptest.NewRequest(http.MethodGet, "/event/day?start=bad", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	assert.Equal(t, serverpkg.ErrInvalidStartPeriod.Error()+"\n", w.Body.String())
}

func TestGetEventsDay_AppError(t *testing.T) {
	app := &mockApp{getEventsDay: func(ctx context.Context, start time.Time) ([]storage.Event, error) {
		_ = ctx
		_ = start
		return nil, errors.New("boom")
	}}
	logger := logger.New("info", os.Stdout, false)
	server := NewServerHTTP("localhost", 8080, logger, app)

	req := httptest.NewRequest(http.MethodGet, "/event/day?start=2025-01-01T00:00:00Z", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
	assert.Equal(t, serverpkg.ErrEventRetrieval.Error()+"\n", w.Body.String())
}

func TestGetEventsDay_Response(t *testing.T) {
	ev := storage.Event{ID: uuid.New(), Title: "Day event"}
	app := &mockApp{getEventsDay: func(ctx context.Context, start time.Time) ([]storage.Event, error) {
		_ = ctx
		_ = start
		return []storage.Event{ev}, nil
	}}
	logger := logger.New("info", os.Stdout, false)
	server := NewServerHTTP("localhost", 8080, logger, app)

	req := httptest.NewRequest(http.MethodGet, "/event/day?start=2025-01-01T00:00:00Z", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	var resp []storage.EventDTO
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, ev.ID, resp[0].ID)
	assert.Equal(t, ev.Title, resp[0].Title)
}
