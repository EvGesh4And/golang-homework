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

// Storage keeps events in memory.
type Storage struct {
	mu        sync.RWMutex
	eventMap  map[uuid.UUID]storage.Event
	intervals IntervalSlice
	logger    *slog.Logger
}

// New creates an in-memory storage instance.
func New(logger *slog.Logger) *Storage {
	return &Storage{
		mu:        sync.RWMutex{},
		eventMap:  make(map[uuid.UUID]storage.Event),
		intervals: IntervalSlice{Intervals: []storage.Interval{}},
		logger:    logger,
	}
}

// CreateEvent adds a new event to storage.
func (s *Storage) setLogCompMeth(ctx context.Context, method string) context.Context {
	ctx = logger.WithLogComponent(ctx, "storage.memory")
	return logger.WithLogMethod(ctx, method)
}

// CreateEvent adds a new event to storage.
func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	ctx = s.setLogCompMeth(ctx, "CreateEvent")
	s.logger.DebugContext(ctx, "attempting to create event")

	if err := ctx.Err(); err != nil {
		return logger.WrapError(ctx, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.eventMap[event.ID]; ok {
		return logger.WrapError(ctx, storage.ErrIDRepeated)
	}
	if !s.intervals.AddIfFree(event.GetInterval()) {
		return logger.WrapError(ctx, storage.ErrDateBusy)
	}

	s.eventMap[event.ID] = event
	s.logger.InfoContext(ctx, "event created successfully")
	return nil
}

// UpdateEvent replaces an existing event.
func (s *Storage) UpdateEvent(ctx context.Context, id uuid.UUID, newEvent storage.Event) error {
	ctx = s.setLogCompMeth(ctx, "UpdateEvent")
	s.logger.DebugContext(ctx, "attempting to update event")

	if err := ctx.Err(); err != nil {
		return logger.WrapError(ctx, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldEvent, ok := s.eventMap[id]
	if !ok {
		return logger.WrapError(ctx, storage.ErrIDNotExist)
	}

	if !s.intervals.Replace(newEvent.GetInterval(), oldEvent.GetInterval()) {
		return logger.WrapError(ctx, storage.ErrDateBusy)
	}

	s.eventMap[id] = newEvent
	s.logger.InfoContext(ctx, "event updated successfully")
	return nil
}

// DeleteEvent removes an event from storage.
func (s *Storage) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	ctx = s.setLogCompMeth(ctx, "DeleteEvent")
	s.logger.DebugContext(ctx, "attempting to delete event")

	if err := ctx.Err(); err != nil {
		return logger.WrapError(ctx, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.eventMap[id]
	if !ok {
		return logger.WrapError(ctx, storage.ErrIDNotExist)
	}

	s.intervals.Remove(event.GetInterval())
	delete(s.eventMap, id)

	s.logger.InfoContext(ctx, "event deleted successfully")
	return nil
}

// GetEventsDay returns events for a day.
func (s *Storage) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, "Day")
}

// GetEventsWeek returns events for a week.
func (s *Storage) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, "Week")
}

// GetEventsMonth returns events for a month.
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

	ctx = s.setLogCompMeth(ctx, fmt.Sprintf("GetEvents%s", period))
	ctx = logger.WithLogStart(ctx, start)

	s.logger.DebugContext(ctx, "attempting to get events for interval")

	if err := ctx.Err(); err != nil {
		return nil, logger.WrapError(ctx, err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	queryInterval := storage.Interval{Start: start, End: start.Add(d)}
	intervals := s.intervals.GetInInterval(queryInterval)

	res := make([]storage.Event, 0, len(intervals))
	for _, inter := range intervals {
		event, ok := s.eventMap[inter.ID]
		if !ok {
			return nil, logger.WrapError(ctx, storage.ErrGetEvents)
		}
		res = append(res, event)
	}

	s.logger.InfoContext(ctx, "events retrieved successfully", "count", len(res))
	return res, nil
}

// Close implements the Storage interface. Nothing to close for memory storage.
func (s *Storage) Close() error {
	return nil // ничего закрывать не нужно
}
