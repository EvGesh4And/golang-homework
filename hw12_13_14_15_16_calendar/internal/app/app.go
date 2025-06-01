package app

import (
	"context"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
)

type App struct {
	logg    Logger
	storage Storage
}

type Logger interface {
	Error(module string, msg string, args ...any)
	Warn(module string, msg string, args ...any)
	Info(module string, msg string, args ...any)
	Debug(module string, msg string, args ...any)
	Printf(msg string, args ...any)
}

type Storage interface {
	CreateEvent(context.Context, storage.Event) error
	UpdateEvent(context.Context, uuid.UUID, storage.Event) error
	DeleteEvent(context.Context, uuid.UUID) error
	GetEventsDay(context.Context, time.Time) ([]storage.Event, error)
	GetEventsWeek(context.Context, time.Time) ([]storage.Event, error)
	GetEventsMonth(context.Context, time.Time) ([]storage.Event, error)
}

func New(logger Logger, storage Storage) *App {
	return &App{
		logg:    logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, event storage.Event) error {
	if err := event.CheckValid(); err != nil {
		a.logg.Error("APP", "CreateEvent некорректное тело события ID %s: %v", event.ID.String(), err)
		return err
	}
	err := a.storage.CreateEvent(ctx, event)
	if err != nil {
		a.logg.Error("APP", "CreateEvent ID %s: storage: %v", event.ID.String(), err)
		return err
	}
	a.logg.Info("APP", "CreateEvent успешно создано событие ID %s", event.ID.String())
	return nil
}

func (a *App) UpdateEvent(ctx context.Context, id uuid.UUID, event storage.Event) error {
	if err := event.CheckValid(); err != nil {
		a.logg.Error("APP", "UpdateEvent некорректное тело события ID %s: %v", event.ID.String(), err)
		return err
	}
	err := a.storage.UpdateEvent(ctx, id, event)
	if err != nil {
		a.logg.Error("APP", "UpdateEvent ID %s: storage: %v", id.String(), err)
		return err
	}
	a.logg.Info("APP", "UpdateEvent успешно обновлено событие ID %s", id.String())
	return nil
}

func (a *App) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	err := a.storage.DeleteEvent(ctx, id)
	if err != nil {
		a.logg.Error("APP", "DeleteEvent событие ID %s: %v", id.String(), err)
		return err
	}
	a.logg.Info("APP", "DeleteEvent успешно удалено событие ID %s", id.String())
	return nil
}

func (a *App) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	events, err := a.storage.GetEventsDay(ctx, start)
	if err != nil {
		a.logg.Error("APP", "GetEventsDay: %v", err)
		return nil, err
	}
	a.logg.Debug("APP", "GetEventsDay: найдено %d событий", len(events))
	return events, nil
}

func (a *App) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	events, err := a.storage.GetEventsWeek(ctx, start)
	if err != nil {
		a.logg.Error("APP", "GetEventsWeek: %v", err)
		return nil, err
	}
	a.logg.Debug("APP", "GetEventsWeek: найдено %d событий", len(events))
	return events, nil
}

func (a *App) GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error) {
	events, err := a.storage.GetEventsMonth(ctx, start)
	if err != nil {
		a.logg.Error("APP", "GetEventsMonth: %v", err)
		return nil, err
	}
	a.logg.Debug("APP", "GetEventsMonth: найдено %d событий", len(events))
	return events, nil
}
