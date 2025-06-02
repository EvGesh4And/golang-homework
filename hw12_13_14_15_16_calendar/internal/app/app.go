package app

import (
	"context"
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
	if err := event.CheckValid(); err != nil {
		a.logger.Error("CreateEvent некорректное тело события ID %s: %v", event.ID.String(), err)
		return err
	}
	err := a.storage.CreateEvent(ctx, event)
	if err != nil {
		a.logger.Error("CreateEvent ID %s: storage: %v", event.ID.String(), err)
		return err
	}
	a.logger.Info("CreateEvent успешно создано событие ID %s", event.ID.String(), nil)
	return nil
}

func (a *App) UpdateEvent(ctx context.Context, id uuid.UUID, event storage.Event) error {
	if err := event.CheckValid(); err != nil {
		a.logger.Error("APP", "UpdateEvent некорректное тело события ID %s: %v", event.ID.String(), err)
		return err
	}
	err := a.storage.UpdateEvent(ctx, id, event)
	if err != nil {
		a.logger.Error("APP", "UpdateEvent ID %s: storage: %v", id.String(), err)
		return err
	}
	a.logger.Info("APP", "UpdateEvent успешно обновлено событие ID %s", id.String())
	return nil
}

func (a *App) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	err := a.storage.DeleteEvent(ctx, id)
	if err != nil {
		a.logger.Error("APP", "DeleteEvent событие ID %s: %v", id.String(), err)
		return err
	}
	a.logger.Info("APP", "DeleteEvent успешно удалено событие ID %s", id.String())
	return nil
}

func (a *App) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	events, err := a.storage.GetEventsDay(ctx, start)
	if err != nil {
		a.logger.Error("APP", "GetEventsDay: %v", err)
		return nil, err
	}
	a.logger.Debug("APP", "GetEventsDay: найдено %d событий", len(events))
	return events, nil
}

func (a *App) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	events, err := a.storage.GetEventsWeek(ctx, start)
	if err != nil {
		a.logger.Error("APP", "GetEventsWeek: %v", err)
		return nil, err
	}
	a.logger.Debug("APP", "GetEventsWeek: найдено %d событий", len(events))
	return events, nil
}

func (a *App) GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error) {
	events, err := a.storage.GetEventsMonth(ctx, start)
	if err != nil {
		a.logger.Error("APP", "GetEventsMonth: %v", err)
		return nil, err
	}
	a.logger.Debug("APP", "GetEventsMonth: найдено %d событий", len(events))
	return events, nil
}
