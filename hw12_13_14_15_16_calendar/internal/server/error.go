package server

import "errors"

var (
	// ErrInvalidContentType means request Content-Type is not application/json.
	ErrInvalidContentType = errors.New("Content-Type must be application/json")
	// ErrMissingEvent occurs when request body lacks event data.
	ErrMissingEvent = errors.New("missing event in request")
	// ErrMissingEventID occurs when event ID is not provided.
	ErrMissingEventID = errors.New("missing event ID in request")
	// ErrInvalidEventID indicates an invalid event ID value.
	ErrInvalidEventID = errors.New("invalid event ID")
	// ErrInvalidUserID indicates an invalid user ID value.
	ErrInvalidUserID = errors.New("invalid user ID")
	// ErrInvalidEventData signals incorrect event data.
	ErrInvalidEventData = errors.New("invalid event data")
	// ErrInvalidPeriod denotes an invalid period value.
	ErrInvalidPeriod = errors.New("invalid period")
	// ErrEventRetrieval is returned when events cannot be retrieved.
	ErrEventRetrieval = errors.New("error retrieving events")
	// ErrInvalidStartPeriod indicates start date has invalid format.
	ErrInvalidStartPeriod = errors.New("invalid period start date")
	// ErrCreateEvent reports a failure during event creation.
	ErrCreateEvent = errors.New("error creating event")
	// ErrUpdateEvent reports a failure during event update.
	ErrUpdateEvent = errors.New("error updating event")
	// ErrDeleteEvent reports a failure during event deletion.
	ErrDeleteEvent = errors.New("error deleting event")
)
