package storage

import (
	"time"
)

type Event struct {
	ID          string
	Title       string
	Start       time.Time
	End         time.Time
	Description string
	UserID      string
	TimeBefore  time.Duration
}

type Interval struct {
	ID    string
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
	return Interval{e.ID, e.Start, e.End}
}
