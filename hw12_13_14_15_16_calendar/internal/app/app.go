package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
)

// App orchestrates application use cases.
type App struct {
	storage Storage
	logger  *slog.Logger
}

func (a *App) setLogCompMeth(ctx context.Context, method string) context.Context {
	ctx = logger.WithLogComponent(ctx, "app")
	return logger.WithLogMethod(ctx, method)
}

// Storage defines persistence methods used by App.
type Storage interface {
	CreateEvent(context.Context, storage.Event) error
	UpdateEvent(context.Context, uuid.UUID, storage.Event) error
	DeleteEvent(context.Context, uuid.UUID) error
	GetEventsDay(context.Context, time.Time) ([]storage.Event, error)
	GetEventsWeek(context.Context, time.Time) ([]storage.Event, error)
	GetEventsMonth(context.Context, time.Time) ([]storage.Event, error)
}

// New creates a new App instance.
func New(logger *slog.Logger, storage Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

// CreateEvent validates and stores a new event.
func (a *App) CreateEvent(ctx context.Context, event storage.Event) error {
	ctx = a.setLogCompMeth(ctx, "CreateEvent")
	a.logger.DebugContext(ctx, "attempting to create event")
	if err := event.CheckValid(); err != nil {
		return logger.AddPrefix(ctx, err)
	}
	err := a.storage.CreateEvent(ctx, event)
	if err != nil {
		return logger.AddPrefix(ctx, err)
	}
	a.logger.InfoContext(ctx, "event created successfully")
	return nil
}

// UpdateEvent validates and updates an existing event.
func (a *App) UpdateEvent(ctx context.Context, id uuid.UUID, event storage.Event) error {
	ctx = a.setLogCompMeth(ctx, "UpdateEvent")
	a.logger.DebugContext(ctx, "attempting to update event")
	if err := event.CheckValid(); err != nil {
		return logger.AddPrefix(ctx, err)
	}
	err := a.storage.UpdateEvent(ctx, id, event)
	if err != nil {
		return logger.AddPrefix(ctx, err)
	}
	a.logger.InfoContext(ctx, "event updated successfully")
	return nil
}

// DeleteEvent removes an event by its ID.
func (a *App) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	ctx = a.setLogCompMeth(ctx, "DeleteEvent")
	a.logger.DebugContext(ctx, "attempting to delete event")
	if err := ctx.Err(); err != nil {
		return logger.AddPrefix(ctx, err)
	}
	err := a.storage.DeleteEvent(ctx, id)
	if err != nil {
		return logger.AddPrefix(ctx, err)
	}
	a.logger.InfoContext(ctx, "event deleted successfully")
	return nil
}

// GetEventsDay retrieves events for one day.
func (a *App) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	ctx = a.setLogCompMeth(ctx, "GetEventsDay")
	ctx = logger.WithLogStart(ctx, start)
	a.logger.DebugContext(ctx, "attempting to get events for day")
	events, err := a.storage.GetEventsDay(ctx, start)
	if err != nil {
		return nil, logger.AddPrefix(ctx, err)
	}
	a.logger.InfoContext(ctx, "events retrieved successfully", "count", len(events))
	return events, nil
}

// GetEventsWeek retrieves events for one week.
func (a *App) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	ctx = a.setLogCompMeth(ctx, "GetEventsWeek")
	ctx = logger.WithLogStart(ctx, start)
	a.logger.DebugContext(ctx, "attempting to get events for week")
	events, err := a.storage.GetEventsWeek(ctx, start)
	if err != nil {
		return nil, logger.AddPrefix(ctx, err)
	}
	a.logger.InfoContext(ctx, "events retrieved successfully", "count", len(events))
	return events, nil
}

// GetEventsMonth retrieves events for one month.
func (a *App) GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error) {
	ctx = a.setLogCompMeth(ctx, "GetEventsMonth")
	ctx = logger.WithLogStart(ctx, start)
	a.logger.DebugContext(ctx, "attempting to get events for month")
	events, err := a.storage.GetEventsMonth(ctx, start)
	if err != nil {
		return nil, logger.AddPrefix(ctx, err)
	}
	a.logger.InfoContext(ctx, "events retrieved successfully", "count", len(events))
	return events, nil
}
