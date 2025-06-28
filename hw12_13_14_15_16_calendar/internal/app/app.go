package app

import (
	"context"
	"fmt"
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
	ctx = logger.WithLogMethod(ctx, "CreateEvent")
	a.logger.DebugContext(ctx, "попытка создать событие")
	if err := event.CheckValid(); err != nil {
		return fmt.Errorf("app.CreateEvent: %w", err)
	}
	err := a.storage.CreateEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("app.CreateEvent: %w", err)
	}
	a.logger.InfoContext(ctx, "успешно создано событие")
	return nil
}

// UpdateEvent validates and updates an existing event.
func (a *App) UpdateEvent(ctx context.Context, id uuid.UUID, event storage.Event) error {
	ctx = logger.WithLogMethod(ctx, "UpdateEvent")
	a.logger.DebugContext(ctx, "попытка обновить событие")
	if err := event.CheckValid(); err != nil {
		return fmt.Errorf("app.UpdateEvent: %w", err)
	}
	err := a.storage.UpdateEvent(ctx, id, event)
	if err != nil {
		return fmt.Errorf("app.UpdateEvent: %w", err)
	}
	a.logger.InfoContext(ctx, "успешно обновлено событие")
	return nil
}

// DeleteEvent removes an event by its ID.
func (a *App) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	ctx = logger.WithLogMethod(ctx, "DeleteEvent")
	a.logger.DebugContext(ctx, "попытка удалить событие")
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("storage:memory.CreateEvent: %w", err)
	}
	err := a.storage.DeleteEvent(ctx, id)
	if err != nil {
		return fmt.Errorf("app.DeleteEvent: %w", err)
	}
	a.logger.InfoContext(ctx, "успешно удалено событие")
	return nil
}

// GetEventsDay retrieves events for one day.
func (a *App) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	ctx = logger.WithLogMethod(ctx, "GetEventsDay")
	ctx = logger.WithLogStart(ctx, start)
	a.logger.DebugContext(ctx, "попытка получить события за день")
	events, err := a.storage.GetEventsDay(ctx, start)
	if err != nil {
		return nil, fmt.Errorf("app.GetEventsDay: %w", err)
	}
	a.logger.InfoContext(ctx, "успешно получены события", "count", len(events))
	return events, nil
}

// GetEventsWeek retrieves events for one week.
func (a *App) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	ctx = logger.WithLogMethod(ctx, "GetEventsWeek")
	ctx = logger.WithLogStart(ctx, start)
	a.logger.DebugContext(ctx, "попытка получить события за неделю")
	events, err := a.storage.GetEventsWeek(ctx, start)
	if err != nil {
		return nil, fmt.Errorf("app.GetEventsWeek: %w", err)
	}
	a.logger.InfoContext(ctx, "успешно получены события", "count", len(events))
	return events, nil
}

// GetEventsMonth retrieves events for one month.
func (a *App) GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error) {
	ctx = logger.WithLogMethod(ctx, "GetEventsMonth")
	ctx = logger.WithLogStart(ctx, start)
	a.logger.DebugContext(ctx, "попытка получить события за месяц")
	events, err := a.storage.GetEventsMonth(ctx, start)
	if err != nil {
		return nil, fmt.Errorf("app.GetEventsMonth: %w", err)
	}
	a.logger.InfoContext(ctx, "успешно получены события", "count", len(events))
	return events, nil
}
