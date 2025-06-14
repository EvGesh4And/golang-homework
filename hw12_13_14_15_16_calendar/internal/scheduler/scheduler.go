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
	Publish(body string) error
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
			s.logger.InfoContext(ctx, "Scheduler остановлен")
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

	s.logger.DebugContext(ctx, "попытка опубликовать уведомления")
	s.logger.DebugContext(ctx, "попытка получить уведомления")

	notifications, err := s.storage.GetNotifications(ctx, currTime, s.tick)
	if err != nil {
		s.logger.ErrorContext(ctx, "ошибка получения уведомлений", "error", err)
		return
	}
	s.logger.InfoContext(ctx, "успешно получены уведомления", "count", len(notifications))
	for _, n := range notifications {
		json, err := json.Marshal(n)
		if err != nil {
			s.logger.ErrorContext(ctx, "ошибка сериализации уведомления", "error", err)
			continue
		}
		if err := s.publisher.Publish(string(json)); err != nil {
			s.logger.ErrorContext(ctx, "ошибка публикации уведомления", "error", err)
			continue
		}
		s.logger.InfoContext(ctx, "опубликовано уведомление", "id", n.ID, "title", n.Title)
	}

	s.logger.InfoContext(ctx, "события успешно опубликованы")

	s.logger.DebugContext(ctx, "попытка удалить старые события")
	err = s.storage.DeleteOldEvents(ctx, currTime.Add(-s.EventTTL))
	if err != nil {
		s.logger.ErrorContext(ctx, "ошибка удаления старых событий", "error", err)
		return
	}
}
