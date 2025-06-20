package sqlstorage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/stdlib" //revive:disable:blank-imports
	"github.com/pressly/goose/v3"
)

type Storage struct {
	dsn    string
	db     *sql.DB
	logger *slog.Logger
}

func New(logger *slog.Logger, dsn string) *Storage {
	return &Storage{
		dsn:    dsn,
		logger: logger,
	}
}

func (s *Storage) Connect(ctx context.Context) (err error) {
	s.db, err = sql.Open("pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("создание пула соединений: %w", err)
	}

	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("установка соединения: %w", err)
	}

	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func (s *Storage) Migrate(migrate string) (err error) {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("установка диалекта: %w", err)
	}

	if err := goose.Up(s.db, migrate); err != nil {
		return fmt.Errorf("ошибка миграции: %w", err)
	}

	return nil
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	ctx = logger.WithLogMethod(ctx, "CreateEvent")
	s.logger.DebugContext(ctx, "попытка создать событие")

	query := `
        INSERT INTO events (id, title, description, user_id, start_time, end_time, time_before)
        VALUES ($1, $2, $3, $4, $5, $6, make_interval(secs => $7))
    `

	_, err := s.db.ExecContext(ctx, query,
		event.ID,
		event.Title,
		event.Description,
		event.UserID,
		event.Start,
		event.End,
		int64(event.TimeBefore.Seconds()),
	)
	if err != nil {
		return logger.WrapError(ctx, fmt.Errorf("storage:sql.CreateEvent: %w", err))
	}
	s.logger.InfoContext(ctx, "успешно создано событие")
	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, id uuid.UUID, newEvent storage.Event) error {
	ctx = logger.WithLogMethod(ctx, "UpdateEvent")
	s.logger.DebugContext(ctx, "попытка обновить событие")

	query := `
        UPDATE events
        SET title = $1, description = $2, user_id = $3, start_time = $4,
		end_time = $5, time_before = make_interval(secs => $6)
        WHERE id = $7
    `

	_, err := s.db.ExecContext(ctx, query,
		newEvent.Title,
		newEvent.Description,
		newEvent.UserID,
		newEvent.Start,
		newEvent.End,
		int64(newEvent.TimeBefore.Seconds()),
		id,
	)
	if err != nil {
		return logger.WrapError(ctx, fmt.Errorf("storage:sql.UpdateEvent: %w", err))
	}
	s.logger.InfoContext(ctx, "успешно обновлено событие")
	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	ctx = logger.WithLogMethod(ctx, "DeleteEvent")

	s.logger.DebugContext(ctx, "попытка удалить событие")

	query := `
        DELETE FROM events
        WHERE id = $1
    `

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return logger.WrapError(ctx, fmt.Errorf("storage:sql.DeleteEvent: %w", err))
	}
	s.logger.InfoContext(ctx, "успешно удалено событие")
	return nil
}

func (s *Storage) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, "Day")
}

func (s *Storage) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, "Week")
}

func (s *Storage) GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, "Month")
}

func (s *Storage) getEvents(ctx context.Context, start time.Time, period string) ([]storage.Event, error) {
	var d time.Duration
	switch period {
	case "Day":
		d = time.Hour * 24
	case "Week":
		d = time.Hour * 24 * 7
	case "Month":
		d = time.Hour * 24 * 30
	}

	ctx = logger.WithLogMethod(ctx, fmt.Sprintf("GetEvents%s", period))
	ctx = logger.WithLogStart(ctx, start)

	s.logger.DebugContext(ctx, "попытка получить события за интервал")

	query := `
        SELECT id, title, description, user_id, start_time, end_time, time_before
        FROM events
        WHERE start_time <= $2 AND end_time >= $1
    `

	rows, err := s.db.QueryContext(ctx, query, start, start.Add(d))
	if err != nil {
		return nil, logger.WrapError(ctx, fmt.Errorf("storage:sql.GetEvents%s: %w", period, err))
	}
	defer rows.Close()

	var events []storage.Event
	for rows.Next() {
		var event storage.Event
		var intervalStr string
		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.UserID,
			&event.Start,
			&event.End,
			&intervalStr,
		); err != nil {
			return nil, logger.WrapError(ctx, fmt.Errorf("storage:sql.GetEvents%s: %w", period, err))
		}

		dur, err := parsePostgresInterval(intervalStr)
		if err != nil {
			return nil, logger.WrapError(ctx, fmt.Errorf("storage:sql.GetEvents%s: %w", period, err))
		}
		event.TimeBefore = dur

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, logger.WrapError(ctx, fmt.Errorf("storage:sql.GetEvents%s: %w", period, err))
	}

	s.logger.InfoContext(ctx, "успешно получены события", "count", len(events))

	return events, nil
}

func (s *Storage) GetNotifications(
	ctx context.Context,
	currTime time.Time,
	tick time.Duration,
) ([]storage.Notification, error) {
	ctx = logger.WithLogMethod(ctx, "GetNotifications")
	ctx = logger.WithLogStart(ctx, currTime)

	s.logger.DebugContext(ctx, "попытка получить события за интервал")

	query := `
        SELECT id, title, start_time, user_id
        FROM events
        WHERE start_time - time_before <= $2 AND start_time - time_before >= $1
    `

	rows, err := s.db.QueryContext(ctx, query, currTime, currTime.Add(tick))
	if err != nil {
		return nil, logger.WrapError(ctx, fmt.Errorf("storage:sql.GetNotifications: %w", err))
	}
	defer rows.Close()

	var notifications []storage.Notification
	for rows.Next() {
		var notification storage.Notification
		if err := rows.Scan(
			&notification.ID,
			&notification.Title,
			&notification.Start,
			&notification.UserID,
		); err != nil {
			return nil, logger.WrapError(ctx, fmt.Errorf("storage:sql.GetNotifications: %w", err))
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, logger.WrapError(ctx, fmt.Errorf("storage:sql.GetNotifications: %w", err))
	}

	s.logger.InfoContext(ctx, "успешно получены уведомления", "count", len(notifications))

	return notifications, nil
}

func (s *Storage) DeleteOldEvents(ctx context.Context, delTime time.Time) error {
	ctx = logger.WithLogMethod(ctx, "DeleteOldEvents")
	ctx = logger.WithLogStart(ctx, delTime)

	s.logger.DebugContext(ctx, "попытка удалить старые события")

	query := `
        DELETE FROM events
        WHERE end_time < $1
    `

	res, err := s.db.ExecContext(ctx, query, delTime)
	if err != nil {
		return logger.WrapError(ctx, fmt.Errorf("storage:sql.DeleteOldEvents: %w", err))
	}
	count, err := res.RowsAffected()
	if err != nil {
		return logger.WrapError(ctx, fmt.Errorf("storage:sql.DeleteOldEvents: %w", err))
	}
	if count > 0 {
		s.logger.InfoContext(ctx, "успешно удалены старые события", "count", count)
	} else {
		s.logger.InfoContext(ctx, "нет старых событий для удаления")
	}
	return nil
}
