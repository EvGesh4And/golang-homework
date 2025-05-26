package storage

import (
	"errors"
	"time"
)

var (
	ErrEventValidTime     = errors.New("ошибка в дате и времени события")
	ErrEventValidDuration = errors.New("ошибка в длительности события")
)

type IDEvent string

type Event struct {
	IDEvent
	Title       string
	Time        time.Time
	Duration    time.Duration
	Description string
	UserID      string
	TimeBefore  time.Duration
}

func (e Event) CheckValid() error {
	if e.Time.Before(time.Now()) {
		return ErrEventValidTime
	}
	if e.Duration < 0 {
		return ErrEventValidDuration
	}
	
}
