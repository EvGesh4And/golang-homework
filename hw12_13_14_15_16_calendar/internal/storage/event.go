package storage

import (
	"time"
)

type IDEvent string

type Event struct {
	IDEvent
	Title       string
	Start       time.Time
	End         time.Time
	Description string
	UserID      string
	TimeBefore  time.Duration
}

type Interval struct {
	IDEvent
	Start time.Time
	End   time.Time
}

func (e Event) CheckValid() error {
	if e.Start.Before(time.Now()) {
		return ErrEventValidStart
	}
	if e.End.Before(e.Start) {
		return ErrEventValidEnd
	}
	if e.TimeBefore < 0 {
		return ErrEventValidBefore
	}
	return nil
}

func (e Event) GetInterval() Interval {
	return Interval{e.IDEvent, e.Start, e.End}
}
