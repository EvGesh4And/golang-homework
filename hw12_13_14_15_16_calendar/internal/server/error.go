package server

import "errors"

var (
	ErrInvalidContentType = errors.New("Content-Type должен быть application/json")
	ErrMissingEventID     = errors.New("отсутствует ID события в запросе")
	ErrInvalidEventID     = errors.New("некорректный ID события")
	ErrInvalidEventData   = errors.New("некорректные данные события")
	ErrInvalidPeriod      = errors.New("некорректный период")
	ErrEventRetrieval     = errors.New("ошибка при получении событий")
	ErrInvalidStartPeriod = errors.New("некорректная дата начала периода")
	ErrCreateEvent        = errors.New("ошибка при создании события")
	ErrUpdateEvent        = errors.New("ошибка при обновлении события")
	ErrDeleteEvent        = errors.New("ошибка при удалении события")
)
