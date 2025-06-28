package server

import "errors"

var (
	// ErrInvalidContentType means request Content-Type is not application/json.
	ErrInvalidContentType = errors.New("Content-Type должен быть application/json")
	// ErrMissingEvent occurs when request body lacks event data.
	ErrMissingEvent = errors.New("отсутствует событие в запросе")
	// ErrMissingEventID occurs when event ID is not provided.
	ErrMissingEventID = errors.New("отсутствует ID события в запросе")
	// ErrInvalidEventID indicates an invalid event ID value.
	ErrInvalidEventID = errors.New("некорректный ID события")
	// ErrInvalidUserID indicates an invalid user ID value.
	ErrInvalidUserID = errors.New("некорректный ID пользователя")
	// ErrInvalidEventData signals incorrect event data.
	ErrInvalidEventData = errors.New("некорректные данные события")
	// ErrInvalidPeriod denotes an invalid period value.
	ErrInvalidPeriod = errors.New("некорректный период")
	// ErrEventRetrieval is returned when events cannot be retrieved.
	ErrEventRetrieval = errors.New("ошибка при получении событий")
	// ErrInvalidStartPeriod indicates start date has invalid format.
	ErrInvalidStartPeriod = errors.New("некорректная дата начала периода")
	// ErrCreateEvent reports a failure during event creation.
	ErrCreateEvent = errors.New("ошибка при создании события")
	// ErrUpdateEvent reports a failure during event update.
	ErrUpdateEvent = errors.New("ошибка при обновлении события")
	// ErrDeleteEvent reports a failure during event deletion.
	ErrDeleteEvent = errors.New("ошибка при удалении события")
)
