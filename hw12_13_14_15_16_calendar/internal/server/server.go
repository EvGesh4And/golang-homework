package server

import (
	"context"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
)

// Application defines business logic used by HTTP and gRPC servers.
type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, id uuid.UUID, event storage.Event) error
	DeleteEvent(ctx context.Context, id uuid.UUID) error
	GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error)
	GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error)
	GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error)
}
