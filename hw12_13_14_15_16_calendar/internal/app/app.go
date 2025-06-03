package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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
	a.logger.Debug("попытка создать событие", "method", "CreateEvent", "eventID", event.ID.String(), "userID", event.UserID.String(), "event", event)
	if err := event.CheckValid(); err != nil {
		return fmt.Errorf("app.CreateEvent: некорректное тело события: %w", err)
	}
	err := a.storage.CreateEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("app.CreateEvent: %w", err)
	}
	a.logger.Info("успешно создано событие", "method", "CreateEvent", "eventID", event.ID.String())
	return nil
}

func (a *App) UpdateEvent(ctx context.Context, id uuid.UUID, event storage.Event) error {
	a.logger.Debug("попытка обновить событие", "method", "UpdateEvent", "eventID", id.String(),
		"userID", event.UserID.String(), "event", event)
	if err := event.CheckValid(); err != nil {
		return fmt.Errorf("app.UpdateEvent: некорректное тело события: %w", err)
	}
	err := a.storage.UpdateEvent(ctx, id, event)
	if err != nil {
		return fmt.Errorf("app.UpdateEvent: %w", err)
	}
	a.logger.Info("успешно обновлено событие", "method", "UpdateEvent", "eventID", event.ID.String())
	return nil
}

func (a *App) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	a.logger.Debug("попытка удалить событие", "method", "DeleteEvent", "eventID", id.String())
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("storage:memory.CreateEvent: %w", err)
	}
	err := a.storage.DeleteEvent(ctx, id)
	if err != nil {
		return fmt.Errorf("app.DeleteEvent: %w", err)
	}
	a.logger.Info("успешно удалено событие", "method", "DeleteEvent", "eventID", id.String())
	return nil
}

func (a *App) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	a.logger.Debug("попытка получить события за день", "method", "GetEventsDay", "start", start.Format(time.RFC3339))
	events, err := a.storage.GetEventsDay(ctx, start)
	if err != nil {
		return nil, fmt.Errorf("app.GetEventsDay: %w", err)
	}
	a.logger.Info("успешно получены события", "method", "GetEventsDay", "count", len(events))
	return events, nil
}

func (a *App) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	a.logger.Debug("попытка получить события за неделю", "method", "GetEventsWeek", "start", start.Format(time.RFC3339))
	events, err := a.storage.GetEventsWeek(ctx, start)
	if err != nil {
		return nil, fmt.Errorf("app.GetEventsWeek: %w", err)
	}
	a.logger.Info("успешно получены события", "method", "GetEventsWeek", "count", len(events))
	return events, nil
}

func (a *App) GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error) {
	a.logger.Debug("попытка получить события за месяц", "method", "GetEventsMonth", "start", start.Format(time.RFC3339))
	events, err := a.storage.GetEventsMonth(ctx, start)
	if err != nil {
		return nil, fmt.Errorf("app.GetEventsMonth: %w", err)
	}
	a.logger.Info("успешно получены события", "method", "GetEventsMonth", "count", len(events))
	return events, nil
}
