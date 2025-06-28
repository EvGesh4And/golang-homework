package storage

import (
	"errors"
	"fmt"
)

// Invalid event fields.
type ErrInvalidEvent struct {
	Field   string
	Message string
}

func (e *ErrInvalidEvent) Error() string {
	return fmt.Sprintf("invalid field %q: %s", e.Field, e.Message)
}

var (
	// Error with ID.
	ErrIDRepeated = errors.New("event with this ID already exists")
	ErrIDNotExist = errors.New("event with this ID does not exist")
	// Error with time intervals.
	ErrDateBusy = errors.New("time slot already occupied by another event")
	// Error retrieving events in interval.
	ErrGetEvents = errors.New("error retrieving list of events")
)
