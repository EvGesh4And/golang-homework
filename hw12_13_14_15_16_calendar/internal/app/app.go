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

type App struct {
	storage Storage
	logger  *slog.Logger
}

type Storage interface {
	CreateEvent(context.Context, storage.Event) error
	UpdateEvent(context.Context, uuid.UUID, storage.Event) error
	DeleteEvent(context.Context, uuid.UUID) error
	GetEventsDay(context.Context, time.Time) ([]storage.Event, error)
	GetEventsWeek(context.Context, time.Time) ([]storage.Event, error)
	GetEventsMonth(context.Context, time.Time) ([]storage.Event, error)
}

func New(logger *slog.Logger, storage Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, event storage.Event) error {
	ctx = logger.WithLogMethod(ctx, "CreateEvent")
	a.logger.DebugContext(ctx, "attempting to create event")
	if err := event.CheckValid(); err != nil {
		return fmt.Errorf("app.CreateEvent: %w", err)
	}
	err := a.storage.CreateEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("app.CreateEvent: %w", err)
	}
	a.logger.InfoContext(ctx, "event successfully created")
	return nil
}

func (a *App) UpdateEvent(ctx context.Context, id uuid.UUID, event storage.Event) error {
	ctx = logger.WithLogMethod(ctx, "UpdateEvent")
	a.logger.DebugContext(ctx, "attempting to update event")
	if err := event.CheckValid(); err != nil {
		return fmt.Errorf("app.UpdateEvent: %w", err)
	}
	err := a.storage.UpdateEvent(ctx, id, event)
	if err != nil {
		return fmt.Errorf("app.UpdateEvent: %w", err)
	}
	a.logger.InfoContext(ctx, "event successfully updated")
	return nil
}

func (a *App) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	ctx = logger.WithLogMethod(ctx, "DeleteEvent")
	a.logger.DebugContext(ctx, "attempting to delete event")
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("storage:memory.CreateEvent: %w", err)
	}
	err := a.storage.DeleteEvent(ctx, id)
	if err != nil {
		return fmt.Errorf("app.DeleteEvent: %w", err)
	}
	a.logger.InfoContext(ctx, "event successfully deleted")
	return nil
}

func (a *App) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	ctx = logger.WithLogMethod(ctx, "GetEventsDay")
	ctx = logger.WithLogStart(ctx, start)
	a.logger.DebugContext(ctx, "attempting to get day's events")
	events, err := a.storage.GetEventsDay(ctx, start)
	if err != nil {
		return nil, fmt.Errorf("app.GetEventsDay: %w", err)
	}
	a.logger.InfoContext(ctx, "events successfully retrieved", "count", len(events))
	return events, nil
}

func (a *App) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	ctx = logger.WithLogMethod(ctx, "GetEventsWeek")
	ctx = logger.WithLogStart(ctx, start)
	a.logger.DebugContext(ctx, "attempting to get week's events")
	events, err := a.storage.GetEventsWeek(ctx, start)
	if err != nil {
		return nil, fmt.Errorf("app.GetEventsWeek: %w", err)
	}
	a.logger.InfoContext(ctx, "events successfully retrieved", "count", len(events))
	return events, nil
}

func (a *App) GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error) {
	ctx = logger.WithLogMethod(ctx, "GetEventsMonth")
	ctx = logger.WithLogStart(ctx, start)
	a.logger.DebugContext(ctx, "attempting to get month's events")
	events, err := a.storage.GetEventsMonth(ctx, start)
	if err != nil {
		return nil, fmt.Errorf("app.GetEventsMonth: %w", err)
	}
	a.logger.InfoContext(ctx, "events successfully retrieved", "count", len(events))
	return events, nil
}
