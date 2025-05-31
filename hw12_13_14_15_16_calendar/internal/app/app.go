package app

import (
	"context"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
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
	UpdateEvent(context.Context, string, storage.Event) error
	DeleteEvent(context.Context, string) error
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
	err := a.storage.CreateEvent(ctx, event)
	if err != nil {
		a.logg.Error("CreateEvent", "событие ID: %s: %v", event.ID, err)
		return err
	}
	a.logg.Info("CreateEvent", "успешно создано событие ID: %s", event.ID)
	return nil
}

func (a *App) UpdateEvent(ctx context.Context, id string, event storage.Event) error {
	err := a.storage.UpdateEvent(ctx, id, event)
	if err != nil {
		a.logg.Error("UpdateEvent", "событие ID: %s: %v", id, err)
		return err
	}
	a.logg.Info("UpdateEvent", "успешно обновлено событие ID: %s", id)
	return nil
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	err := a.storage.DeleteEvent(ctx, id)
	if err != nil {
		a.logg.Error("DeleteEvent", "событие ID: %s: %v", id, err)
		return err
	}
	a.logg.Info("DeleteEvent", "успешно удалено событие ID: %s", id)
	return nil
}

func (a *App) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	events, err := a.storage.GetEventsDay(ctx, start)
	if err != nil {
		a.logg.Error("GetEventsDay", " %v", err)
		return nil, err
	}
	a.logg.Debug("GetEventsDay", "найдено %d событий", len(events))
	return events, nil
}

func (a *App) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	events, err := a.storage.GetEventsWeek(ctx, start)
	if err != nil {
		a.logg.Error("GetEventsWeek", "%v", err)
		return nil, err
	}
	a.logg.Debug("GetEventsWeek", "найдено %d событий", len(events))
	return events, nil
}

func (a *App) GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error) {
	events, err := a.storage.GetEventsMonth(ctx, start)
	if err != nil {
		a.logg.Error("GetEventsMonth", "%v", err)
		return nil, err
	}
	a.logg.Debug("GetEventsMonth", "найдено %d событий", len(events))
	return events, nil
}
