package storage

import "errors"

var (
	ErrValidEvent      = errors.New("событие некорректно")
	ErrIDEventRepeated = errors.New("событие с таким ID уже есть в хранилище")
	ErrIDEventNotExist = errors.New("события с таким ID не существует")
)
