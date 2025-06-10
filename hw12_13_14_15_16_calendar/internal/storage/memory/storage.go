package memorystorage

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
)

type Storage struct {
	mu        sync.RWMutex
	eventMap  map[uuid.UUID]storage.Event
	intervals IntervalSlice
	logger    *slog.Logger
}

func New(logger *slog.Logger) *Storage {
	return &Storage{
		mu:        sync.RWMutex{},
		eventMap:  make(map[uuid.UUID]storage.Event),
		intervals: IntervalSlice{Intervals: []storage.Interval{}},
		logger:    logger,
	}
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	ctx = logger.WithLogMethod(ctx, "CreateEvent")
	s.logger.DebugContext(ctx, "попытка создать событие")

	if err := ctx.Err(); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("storage:memory.CreateEvent: %w", err))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.eventMap[event.ID]; ok {
		return logger.WrapError(ctx, fmt.Errorf("storage:memory.CreateEvent: %w", storage.ErrIDRepeated))
	}
	if !s.intervals.AddIfFree(event.GetInterval()) {
		return logger.WrapError(ctx, fmt.Errorf("storage:memory.CreateEvent: %w", storage.ErrDateBusy))
	}

	s.eventMap[event.ID] = event
	s.logger.InfoContext(ctx, "успешно создано событие")
	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, id uuid.UUID, newEvent storage.Event) error {
	s.logger.DebugContext(ctx, "попытка обновить событие")

	if err := ctx.Err(); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("storage:memory.UpdateEvent: %w", err))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldEvent, ok := s.eventMap[id]
	if !ok {
		return logger.WrapError(ctx, fmt.Errorf("storage:memory.UpdateEvent: %w", storage.ErrIDNotExist))
	}

	if !s.intervals.Replace(newEvent.GetInterval(), oldEvent.GetInterval()) {
		return logger.WrapError(ctx, fmt.Errorf("storage:memory.UpdateEvent: %w", storage.ErrDateBusy))
	}

	s.eventMap[id] = newEvent
	s.logger.InfoContext(ctx, "успешно обновлено событие")
	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	s.logger.DebugContext(ctx, "попытка удалить событие")

	if err := ctx.Err(); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("storage:memory.DeleteEvent: %w", err))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.eventMap[id]
	if !ok {
		return logger.WrapError(ctx, fmt.Errorf("storage:memory.DeleteEvent: %w", storage.ErrIDNotExist))
	}

	s.intervals.Remove(event.GetInterval())
	delete(s.eventMap, id)

	s.logger.InfoContext(ctx, "успешно удалено событие")
	return nil
}

func (s *Storage) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, "Day")
}

func (s *Storage) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, "Week")
}

func (s *Storage) GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, "Month")
}

func (s *Storage) getEvents(ctx context.Context, start time.Time, period string) ([]storage.Event, error) {
	var d time.Duration
	switch period {
	case "Day":
		d = time.Hour * 24
	case "Week":
		d = time.Hour * 24 * 7
	case "Month":
		d = time.Hour * 24 * 30
	}

	ctx = logger.WithLogMethod(ctx, fmt.Sprintf("GetEvents%s", period))
	ctx = logger.WithLogStart(ctx, start)

	s.logger.DebugContext(ctx, "попытка получить события за интервал")

	if err := ctx.Err(); err != nil {
		return nil, logger.WrapError(ctx, fmt.Errorf("storage:memory.getEvents: %w", err))
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	queryInterval := storage.Interval{Start: start, End: start.Add(d)}
	intervals := s.intervals.GetInInterval(queryInterval)

	res := make([]storage.Event, 0, len(intervals))
	for _, inter := range intervals {
		event, ok := s.eventMap[inter.ID]
		if !ok {
			return nil, logger.WrapError(ctx, fmt.Errorf("storage:memory.getEvents: %w", storage.ErrGetEvents))
		}
		res = append(res, event)
	}

	s.logger.InfoContext(ctx, "успешно получены события", "count", len(res))
	return res, nil
}
