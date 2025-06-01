package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
)

type Storage struct {
	mu        sync.RWMutex
	eventMap  map[uuid.UUID]storage.Event
	intervals IntervalSlice
}

func New() *Storage {
	return &Storage{
		mu:        sync.RWMutex{},
		eventMap:  make(map[uuid.UUID]storage.Event),
		intervals: IntervalSlice{Intervals: []storage.Interval{}},
	}
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.eventMap[event.ID]; ok {
		return storage.ErrIDRepeated
	}
	if !s.intervals.AddIfFree(event.GetInterval()) {
		return storage.ErrDateBusy
	}

	s.eventMap[event.ID] = event
	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, id uuid.UUID, newEvent storage.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldEvent, ok := s.eventMap[id]
	if !ok {
		return storage.ErrIDNotExist
	}

	if !s.intervals.Replace(newEvent.GetInterval(), oldEvent.GetInterval()) {
		return storage.ErrDateBusy
	}

	s.eventMap[id] = newEvent
	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.eventMap[id]
	if !ok {
		return storage.ErrIDNotExist
	}

	s.intervals.Remove(event.GetInterval())
	delete(s.eventMap, id)
	return nil
}

func (s *Storage) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, time.Hour*24)
}

func (s *Storage) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, time.Hour*24*7)
}

func (s *Storage) GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, time.Hour*24*30)
}

func (s *Storage) getEvents(ctx context.Context, start time.Time, d time.Duration) ([]storage.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	intervals := s.intervals.GetInInterval(storage.Interval{Start: start, End: start.Add(d)})
	res := make([]storage.Event, 0, len(intervals))

	for _, inter := range intervals {
		event, ok := s.eventMap[inter.ID]
		if !ok {
			return nil, storage.ErrGetEvents
		}
		res = append(res, event)
	}

	return res, nil
}
