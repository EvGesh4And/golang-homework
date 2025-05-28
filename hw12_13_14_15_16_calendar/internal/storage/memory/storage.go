package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/EvGesh4And/hw12_13_14_15_calendar/internal/app"
	"github.com/EvGesh4And/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	logg      app.Logger
	mu        sync.RWMutex
	eventMap  map[string]storage.Event
	intervals IntervalSlice
}

func New(logg app.Logger) *Storage {
	return &Storage{
		logg:      logg,
		mu:        sync.RWMutex{},
		eventMap:  make(map[string]storage.Event),
		intervals: IntervalSlice{Intervals: []storage.Interval{}},
	}
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	if err := event.CheckValid(); err != nil {
		s.logg.Error("CreateEvent: некорректное событие: %v", err)
		return err
	}

	if err := ctx.Err(); err != nil {
		s.logg.Error("CreateEvent: операция отменена контекстом")
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.eventMap[event.ID]; ok {
		s.logg.Error("CreateEvent: событие с ID: %s уже существует", event.ID)
		return storage.ErrIDRepeated
	}
	if !s.intervals.AddIfFree(event.GetInterval()) {
		s.logg.Error("CreateEvent: временной интервал занят для ID: %s", event.ID)
		return storage.ErrDateBusy
	}

	s.eventMap[event.ID] = event
	s.logg.Debug("CreateEvent: добавлено событие ID: %s", event.ID)
	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, ID string, newEvent storage.Event) error {
	if err := newEvent.CheckValid(); err != nil {
		s.logg.Error("UpdateEvent: некорректное новое событие: %v", err)
		return err
	}

	if err := ctx.Err(); err != nil {
		s.logg.Error("UpdateEvent: операция отменена контекстом")
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldEvent, ok := s.eventMap[ID]
	if !ok {
		s.logg.Error("UpdateEvent: событие с ID: %s уже существует", ID)
		return storage.ErrIDNotExist
	}

	if !s.intervals.Replace(newEvent.GetInterval(), oldEvent.GetInterval()) {
		s.logg.Error("UpdateEvent: новый временной интервал для ID: %s уже занят", ID)
		return storage.ErrDateBusy
	}

	s.eventMap[ID] = newEvent
	s.logg.Debug("UpdateEvent: событие с ID: %s обновлено", ID)
	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, ID string) error {
	if err := ctx.Err(); err != nil {
		s.logg.Error("DeleteEvent: операция отменена контекстом")
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.eventMap[ID]
	if !ok {
		s.logg.Error("DeleteEvent: события ID: %s не существует", ID)
		return storage.ErrIDNotExist
	}

	s.intervals.Remove(event.GetInterval())
	delete(s.eventMap, ID)
	s.logg.Debug("DeleteEvent: удалено событие ID: %s", ID)
	return nil
}

func (s *Storage) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEventsWithLog(ctx, "GetEventsDay", start, time.Hour*24)
}

func (s *Storage) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEventsWithLog(ctx, "GetEventsWeek", start, time.Hour*24*7)
}

func (s *Storage) GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEventsWithLog(ctx, "GetEventsMonth", start, time.Hour*24*30)
}

func (s *Storage) getEventsWithLog(ctx context.Context, method string, start time.Time, duration time.Duration) ([]storage.Event, error) {
	res := []storage.Event{}

	if err := ctx.Err(); err != nil {
		s.logg.Error("%s: операция отменена контекстом", method)
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	intervals := s.intervals.GetStartedInInterval(storage.Interval{Start: start, End: start.Add(duration)})

	for _, inter := range intervals {
		event, ok := s.eventMap[inter.ID]
		if !ok {
			s.logg.Error("%s: ошибка с событием ID: %s", method, inter.ID)
			return nil, storage.ErrGetEvents
		}
		res = append(res, event)
	}

	s.logg.Debug("%s: найдено %d событий", method, len(res))

	return res, nil
}
