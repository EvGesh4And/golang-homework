package internalhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
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

	logger := logger.New("info", os.Stdout)
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

	logger := logger.New("info", os.Stdout)
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

	logger := logger.New("info", os.Stdout)
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

	logger := logger.New("info", os.Stdout)
	server := NewServerHTTP("localhost", 8080, logger, app)
	req := httptest.NewRequest(http.MethodGet, "/event?start=2025-01-01T00:00:00Z&period=day", nil)
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

	logger := logger.New("info", os.Stdout)
	server := NewServerHTTP("localhost", 8080, logger, app)
	req := httptest.NewRequest(http.MethodGet, "/event?start=2025-01-01T00:00:00Z&period=week", nil)
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

	logger := logger.New("info", os.Stdout)
	server := NewServerHTTP("localhost", 8080, logger, app)
	req := httptest.NewRequest(http.MethodGet, "/event?start=2025-01-01T00:00:00Z&period=month", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}
