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
	return fmt.Sprintf("некорректное поле %q: %s", e.Field, e.Message)
}

var (
	// Ошибка с ID.
	ErrIDRepeated = errors.New("событие с таким ID уже есть в хранилище")
	ErrIDNotExist = errors.New("события с таким ID не существует")
	// Ошибка с временными интервалами.
	ErrDateBusy = errors.New("данное время уже занято другим событием")
	// Ошибка с получем событий в интервале.
	ErrGetEvents = errors.New("ошибка в ходе получения списка событий")
)
