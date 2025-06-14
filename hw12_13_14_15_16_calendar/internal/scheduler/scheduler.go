package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
)

type Storage interface {
	GetNotifications(ctx context.Context, start time.Time, tick time.Duration) ([]storage.Notification, error)
}

type Scheduler struct {
	storage Storage
	tick    time.Duration
	logger  *slog.Logger
}

func NewScheduler(logger *slog.Logger, storage Storage, tick time.Duration) *Scheduler {
	return &Scheduler{
		storage: storage,
		tick:    tick,
		logger:  logger,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ctx = logger.WithLogMethod(ctx, "Start")

	ticker := time.NewTicker(s.tick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
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
	s.logger.DebugContext(ctx, "попытка получить уведомления")

	notifications, err := s.storage.GetNotifications(ctx, currTime, s.tick)
	if err != nil {
		s.logger.ErrorContext(ctx, "ошибка получения уведомлений", "error", err)
		return
	}
	s.logger.InfoContext(ctx, "успешно получены уведомления", "count", len(notifications))
	s.logger.DebugContext(ctx, "попытка опубликовать уведомления")
	for _, n := range notifications {
		s.logger.InfoContext(ctx, "опубликовано уведомление", "id", n.ID, "title", n.Title)
	}
}
