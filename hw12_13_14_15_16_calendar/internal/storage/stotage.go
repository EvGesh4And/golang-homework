package storage

import "time"

type Storage interface {
	AddEvent(Event) error
	UpdateEvent(IDEvent, Event) error
	DeleteEvent(IDEvent) error
	GetEventsDay(time.Time)
	GetEventsWeek(time.Time)
	GetEventsMonth(time.Time)
}
