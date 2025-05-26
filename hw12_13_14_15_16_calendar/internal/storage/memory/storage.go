package memorystorage

import (
	"sync"

	"github.com/EvGesh4And/hw12_13_14_15_calendar/internal/logger"
	"github.com/EvGesh4And/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	logg     *logger.Logger
	mu       sync.RWMutex //nolint:unused
	eventMap map[storage.IDEvent]storage.Event
}

func New(logg *logger.Logger) *Storage {
	return &Storage{logg, sync.RWMutex{}, make(map[storage.IDEvent]storage.Event)}
}

func (s *Storage) AddEvent(event storage.Event) error {
	if _, ok := s.eventMap[event.IDEvent]; ok {
		return storage.ErrIDEventRepeated
	}
	s.eventMap[event.IDEvent] = event
	return nil
}

func (s *Storage) UpdateEvent(idEvent storage.IDEvent, event storage.Event) error {
	if _, ok := s.eventMap[idEvent]; !ok {
		return storage.ErrIDEventNotExist
	}
	s.eventMap[event.IDEvent] = event
	return nil
}
