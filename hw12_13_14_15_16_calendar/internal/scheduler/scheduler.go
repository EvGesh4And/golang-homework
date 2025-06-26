package scheduler

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
)

type Storage interface {
	GetNotifications(ctx context.Context, start time.Time, tick time.Duration) ([]storage.Notification, error)
	DeleteOldEvents(ctx context.Context, before time.Time) error
}

type Publisher interface {
	Publish(ctx context.Context, body string) error
	Shutdown() error
}

type Scheduler struct {
	storage   Storage
	publisher Publisher
	tick      time.Duration
	EventTTL  time.Duration
	logger    *slog.Logger
}

func NewScheduler(logger *slog.Logger, storage Storage, publisher Publisher, cfg NotificationsConf) *Scheduler {
	return &Scheduler{
		storage:   storage,
		publisher: publisher,
		tick:      cfg.Tick,
		EventTTL:  cfg.EventTTL,
		logger:    logger,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ctx = logger.WithLogMethod(ctx, "Start")

	s.logger.InfoContext(ctx, "start scheduler")

	ticker := time.NewTicker(s.tick)
	defer ticker.Stop()

	s.PublishNotifications(ctx)
	for {
		select {
		case <-ctx.Done():
			s.publisher.Shutdown()
			s.logger.InfoContext(ctx, "scheduler stopped")
			return
		case <-ticker.C:
			s.PublishNotifications(ctx)
		}
	}
}

func (s *Scheduler) PublishNotifications(ctx context.Context) {
	ctx = logger.WithLogMethod(ctx, "PublishNotifications")

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
		json, err := json.Marshal(n)
		if err != nil {
			s.logger.ErrorContext(ctx, "Scheduler.PublishNotifications: failed to serialize notification", "error", err)
			continue
		}
		s.logger.DebugContext(ctx, "successfully serialized notification", "id", n.ID)
		s.logger.DebugContext(ctx, "trying to publish notification", "id", n.ID)
		if err := s.publisher.Publish(ctx, string(json)); err != nil {
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
