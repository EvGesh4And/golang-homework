package memorystorage

import (
	"sync"
	"time"

	"github.com/EvGesh4And/hw12_13_14_15_calendar/internal/logger"
	"github.com/EvGesh4And/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	logg      *logger.Logger
	mu        sync.RWMutex
	eventMap  map[storage.IDEvent]storage.Event
	intervals IntervalSlice
}

func New(logg *logger.Logger) *Storage {
	return &Storage{
		logg:      logg,
		mu:        sync.RWMutex{},
		eventMap:  make(map[storage.IDEvent]storage.Event),
		intervals: IntervalSlice{Intervals: []storage.Interval{}},
	}
}

func (s *Storage) AddEvent(event storage.Event) error {
	if err := event.CheckValid(); err != nil {
		s.logg.Error("AddEvent: некорректное событие: %v", err)
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.eventMap[event.IDEvent]; ok {
		s.logg.Error("AddEvent: ID уже существует: %s", event.IDEvent)
		return storage.ErrIDEventRepeated
	}
	if !s.intervals.AddIfFree(event.GetInterval()) {
		s.logg.Error("AddEvent: интервал для ID занят: %s", event.IDEvent)
		return storage.ErrDateBusy
	}

	s.eventMap[event.IDEvent] = event
	s.logg.Debug("AddEvent: added event ID=%s", event.IDEvent)
	return nil
}

func (s *Storage) UpdateEvent(idEvent storage.IDEvent, newEvent storage.Event) error {
	if err := newEvent.CheckValid(); err != nil {
		s.logg.Error("UpdateEvent: invalid new event: %v", err)
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	oldEvent, ok := s.eventMap[idEvent]
	if !ok {
		s.logg.Error("UpdateEvent: ID does not exist: %s", idEvent)
		return storage.ErrIDEventNotExist
	}

	if !s.intervals.Replace(newEvent.GetInterval(), oldEvent.GetInterval()) {
		s.logg.Error("UpdateEvent: interval conflict for ID: %s", idEvent)
		return storage.ErrDateBusy
	}

	s.eventMap[idEvent] = newEvent
	s.logg.Debug("UpdateEvent: updated event ID=%s", idEvent)
	return nil
}

func (s *Storage) DeleteEvent(idEvent storage.IDEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.eventMap[idEvent]
	if !ok {
		s.logg.Error("DeleteEvent: ID does not exist: %s", idEvent)
		return storage.ErrIDEventNotExist
	}

	s.intervals.Remove(event.GetInterval())
	delete(s.eventMap, idEvent)
	s.logg.Debug("DeleteEvent: deleted event ID=%s", idEvent)
	return nil
}

func (s *Storage) GetEventsDay(start time.Time) ([]storage.Event, error) {
	return s.getEventsWithLog("GetEventsDay", start, time.Hour*24)
}

func (s *Storage) GetEventsWeek(start time.Time) ([]storage.Event, error) {
	return s.getEventsWithLog("GetEventsWeek", start, time.Hour*24*7)
}

func (s *Storage) GetEventsMonth(start time.Time) ([]storage.Event, error) {
	return s.getEventsWithLog("GetEventsMonth", start, time.Hour*24*30)
}

func (s *Storage) getEventsWithLog(method string, start time.Time, duration time.Duration) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var res []storage.Event
	intervals := s.intervals.GetStartedInInterval(storage.Interval{Start: start, End: start.Add(duration)})
	for _, inter := range intervals {
		event, ok := s.eventMap[inter.IDEvent]
		if !ok {
			s.logg.Error("%s: missing event for ID: %s", method, inter.IDEvent)
			return nil, storage.ErrGetEvents
		}
		res = append(res, event)
	}

	s.logg.Debug("%s: found %d events", method, len(res))
	return res, nil
}
