package server

import "errors"

var (
	ErrInvalidContentType = errors.New("Content-Type must be application/json")
	ErrMissingEvent       = errors.New("event missing in request")
	ErrMissingEventID     = errors.New("event ID missing in request")
	ErrInvalidEventID     = errors.New("invalid event ID")
	ErrInvalidUserID      = errors.New("invalid user ID")
	ErrInvalidEventData   = errors.New("invalid event data")
	ErrInvalidPeriod      = errors.New("invalid period")
	ErrEventRetrieval     = errors.New("error retrieving events")
	ErrInvalidStartPeriod = errors.New("invalid start date")
	ErrCreateEvent        = errors.New("error creating event")
	ErrUpdateEvent        = errors.New("error updating event")
	ErrDeleteEvent        = errors.New("error deleting event")
)
