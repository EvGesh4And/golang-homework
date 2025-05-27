package storage

import "time"

type Storage interface {
	AddEvent(Event) error
	UpdateEvent(IDEvent, Event) error
	DeleteEvent(IDEvent) error
	GetEventsDay(time.Time) ([]Event, error)
	GetEventsWeek(time.Time) ([]Event, error)
	GetEventsMonth(time.Time) ([]Event, error)
}
