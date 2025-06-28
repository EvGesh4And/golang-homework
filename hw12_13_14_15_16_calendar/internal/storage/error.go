package storage

import (
	"errors"
	"fmt"
)

// Ошибка в полях события.
type ErrInvalidEvent struct {
	Field   string
	Message string
}

func (e *ErrInvalidEvent) Error() string {
	return fmt.Sprintf("invalid field %q: %s", e.Field, e.Message)
}

var (
	// Ошибка с ID.
	ErrIDRepeated = errors.New("event with this ID already exists")
	ErrIDNotExist = errors.New("event with this ID does not exist")
	// Ошибка с временными интервалами.
	ErrDateBusy = errors.New("time slot already occupied by another event")
	// Ошибка с получем событий в интервале.
	ErrGetEvents = errors.New("error retrieving list of events")
)
