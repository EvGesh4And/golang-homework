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
	Close()
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

	ticker := time.NewTicker(s.tick)
	defer ticker.Stop()

	s.PublishNotifications(ctx)
	for {
		select {
		case <-ctx.Done():
			s.publisher.Close()
			s.logger.InfoContext(ctx, "Scheduler stopped")
			return
		case <-ticker.C:
			s.PublishNotifications(ctx)
		}
	}
}

func (s *Scheduler) PublishNotifications(ctx context.Context) {
	currTime := time.Now()
	ctx = logger.WithLogMethod(ctx, "PublishNotifications")
	ctx = logger.WithLogStart(ctx, currTime)

	s.logger.DebugContext(ctx, "attempting to publish notifications")
	s.logger.DebugContext(ctx, "attempting to fetch notifications")

	notifications, err := s.storage.GetNotifications(ctx, currTime, s.tick)
	if err != nil {
		s.logger.ErrorContext(ctx, "error retrieving notifications", "error", err)
		return
	}
	s.logger.InfoContext(ctx, "notifications retrieved successfully", "count", len(notifications))
	for _, n := range notifications {
		json, err := json.Marshal(n)
		if err != nil {
			s.logger.ErrorContext(ctx, "error serializing notification", "error", err)
			continue
		}
		if err := s.publisher.Publish(ctx, string(json)); err != nil {
			s.logger.ErrorContext(ctx, "error publishing notification", "error", err)
			continue
		}
		s.logger.InfoContext(ctx, "notification published", "id", n.ID, "title", n.Title)
	}

	s.logger.InfoContext(ctx, "events successfully published")

	s.logger.DebugContext(ctx, "attempting to delete old events")
	err = s.storage.DeleteOldEvents(ctx, currTime.Add(-s.EventTTL))
	if err != nil {
		s.logger.ErrorContext(ctx, "error deleting old events", "error", err)
		return
	}
}
