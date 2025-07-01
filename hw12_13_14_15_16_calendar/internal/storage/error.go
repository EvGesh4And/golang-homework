package storage

import (
	"errors"
	"fmt"
)

// ErrInvalidEvent describes invalid fields of an event.
type ErrInvalidEvent struct {
	Field   string
	Message string
}

func (e *ErrInvalidEvent) Error() string {
	return fmt.Sprintf("invalid field %q: %s", e.Field, e.Message)
}

var (
	// Ошибка с ID.
	ErrIDRepeated = errors.New("event with such ID already exists in storage")
	ErrIDNotExist = errors.New("event with such ID does not exist")
	// Ошибка с временными интервалами.
	ErrDateBusy = errors.New("this time is already occupied by another event")
	// Ошибка с получем событий в интервале.
	ErrGetEvents = errors.New("error while retrieving events list")
)
