package scheduler

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
)

// Storage provides access to event data needed by the scheduler.
type Storage interface {
	GetNotifications(ctx context.Context, start time.Time, tick time.Duration) ([]storage.Notification, error)
	DeleteOldEvents(ctx context.Context, before time.Time) error
}

// Publisher sends notifications about upcoming events.
type Publisher interface {
	Publish(ctx context.Context, body string) error
	Shutdown() error
}

// Scheduler periodically publishes event notifications.
type Scheduler struct {
	storage   Storage
	publisher Publisher
	tick      time.Duration
	EventTTL  time.Duration
	logger    *slog.Logger
}

func (s *Scheduler) setLogCompMeth(ctx context.Context, method string) context.Context {
	ctx = logger.WithLogComponent(ctx, "scheduler")
	return logger.WithLogMethod(ctx, method)
}

// NewScheduler creates a new Scheduler instance.
func NewScheduler(logger *slog.Logger, storage Storage, publisher Publisher, cfg NotificationsConf) *Scheduler {
	return &Scheduler{
		storage:   storage,
		publisher: publisher,
		tick:      cfg.Tick,
		EventTTL:  cfg.EventTTL,
		logger:    logger,
	}
}

// Start runs the scheduler loop until the context is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	ctx = s.setLogCompMeth(ctx, "Start")

	s.logger.InfoContext(ctx, "start scheduler")

	ticker := time.NewTicker(s.tick)
	defer ticker.Stop()

	s.PublishNotifications(ctx)
	for {
		select {
		case <-ctx.Done():
			if err := s.publisher.Shutdown(); err != nil {
				s.logger.ErrorContext(ctx, "failed to shutdown publisher", "error", err)
				return
			}
			s.logger.InfoContext(ctx, "scheduler stopped")
			return
		case <-ticker.C:
			s.PublishNotifications(ctx)
		}
	}
}

// PublishNotifications sends notifications and removes old events.
func (s *Scheduler) PublishNotifications(ctx context.Context) {
	ctx = s.setLogCompMeth(ctx, "PublishNotifications")

	currTime := time.Now()
	ctx = logger.WithLogStart(ctx, currTime)

	s.logger.DebugContext(ctx, "trying to publish notifications")
	s.logger.DebugContext(ctx, "trying to get notifications")

	notifications, err := s.storage.GetNotifications(ctx, currTime, s.tick)
	if err != nil {
		s.logger.ErrorContext(ctx, "Scheduler.PublishNotifications: failed to get notifications", "error", err)
		return
	}
	s.logger.InfoContext(ctx, "successfully got notifications", "count", len(notifications))
	for _, n := range notifications {
		s.logger.DebugContext(ctx, "trying to serialize notification", "id", n.ID)
		jsonData, err := json.Marshal(n)
		if err != nil {
			s.logger.ErrorContext(ctx, "Scheduler.PublishNotifications: failed to serialize notification", "error", err)
			continue
		}
		s.logger.DebugContext(ctx, "successfully serialized notification", "id", n.ID)
		s.logger.DebugContext(ctx, "trying to publish notification", "id", n.ID)
		if err := s.publisher.Publish(ctx, string(jsonData)); err != nil {
			s.logger.ErrorContext(ctx, "Scheduler.PublishNotifications: failed to publish notification", "error", err)
			continue
		}
		s.logger.InfoContext(ctx, "Scheduler.PublishNotifications: notification published", "id", n.ID, "title", n.Title)
	}

	s.logger.InfoContext(ctx, "Scheduler.PublishNotifications: events successfully published")

	s.logger.DebugContext(ctx, "trying to delete old events")
	err = s.storage.DeleteOldEvents(ctx, currTime.Add(-s.EventTTL))
	if err != nil {
		s.logger.ErrorContext(ctx, "Scheduler.PublishNotifications: failed to delete old events", "error", err)
	}
	s.logger.InfoContext(ctx, "Scheduler.PublishNotifications: old events deleted")
}
