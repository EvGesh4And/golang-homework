package memorystorage

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

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
	s.logger.Debug("попытка создать событие", "method", "CreateEvent", "eventID",
		event.ID.String(), "userID", event.UserID.String(), "event", event)

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("storage:memory.CreateEvent: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.eventMap[event.ID]; ok {
		return fmt.Errorf("storage:memory.CreateEvent: %w", storage.ErrIDRepeated)
	}
	if !s.intervals.AddIfFree(event.GetInterval()) {
		return fmt.Errorf("storage:memory.CreateEvent: %w", storage.ErrDateBusy)
	}

	s.eventMap[event.ID] = event
	s.logger.Info("успешно создано событие", "method", "CreateEvent",
		"eventID", event.ID.String(), "userID", event.UserID.String())
	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, id uuid.UUID, newEvent storage.Event) error {
	s.logger.Debug("попытка обновить событие", "method", "UpdateEvent",
		"eventID", id.String(), "newUserID", newEvent.UserID.String(), "newEvent", newEvent)

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("storage:memory.UpdateEvent: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldEvent, ok := s.eventMap[id]
	if !ok {
		s.logger.Info("событие для обновления не найдено",
			"method", "UpdateEvent", "eventID", id.String())
		return fmt.Errorf("storage:memory.UpdateEvent: %w", storage.ErrIDNotExist)
	}

	if !s.intervals.Replace(newEvent.GetInterval(), oldEvent.GetInterval()) {
		return fmt.Errorf("storage:memory.UpdateEvent: %w", storage.ErrDateBusy)
	}

	s.eventMap[id] = newEvent
	s.logger.Info("успешно обновлено событие", "method", "UpdateEvent",
		"eventID", id.String(), "userID", newEvent.UserID.String())
	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	s.logger.Debug("попытка удалить событие", "method", "DeleteEvent", "eventID", id.String())

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("storage:memory.DeleteEvent: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.eventMap[id]
	if !ok {
		return fmt.Errorf("storage:memory.DeleteEvent: %w", storage.ErrIDNotExist)
	}

	s.intervals.Remove(event.GetInterval())
	delete(s.eventMap, id)

	s.logger.Info("успешно удалено событие", "method", "DeleteEvent",
		"eventID", id.String(), "userID", event.UserID.String())
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

	s.logger.Debug(
		"попытка получить события за интервал",
		"method", fmt.Sprintf("GetEvents%s", period),
		"start", start.Format(time.RFC3339),
	)

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("storage:memory.getEvents: %w", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	queryInterval := storage.Interval{Start: start, End: start.Add(d)}
	intervals := s.intervals.GetInInterval(queryInterval)

	res := make([]storage.Event, 0, len(intervals))
	for _, inter := range intervals {
		event, ok := s.eventMap[inter.ID]
		if !ok {
			return nil, fmt.Errorf("storage:memory.getEvents: %w", storage.ErrGetEvents)
		}
		res = append(res, event)
	}

	s.logger.Info(
		"успешно получены события",
		"method", fmt.Sprintf("GetEvents%s", period),
		"count", len(res),
		"start", start.Format(time.RFC3339),
	)
	return res, nil
}
