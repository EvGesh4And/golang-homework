package storage

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          uuid.UUID     `json:"id"`
	Title       string        `json:"title"`
	Start       time.Time     `json:"start"`
	End         time.Time     `json:"end"`
	Description string        `json:"description"`
	UserID      uuid.UUID     `json:"userId"`
	TimeBefore  time.Duration `json:"timeBefore"`
}

type Interval struct {
	ID    uuid.UUID
	Start time.Time
	End   time.Time
}

func (e Event) CheckValid() error {
	if e.ID == uuid.Nil {
		return ErrEventValidID
	}
	if e.UserID == uuid.Nil {
		return ErrEventValidUserID
	}
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
