package sqlstorage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/stdlib" //revive:disable:blank-imports
	"github.com/pressly/goose/v3"
)

// Storage works with PostgreSQL to persist events.
type Storage struct {
	dsn    string
	db     *sql.DB
	logger *slog.Logger
}

// New creates a new SQL storage with the given DSN.
func New(logger *slog.Logger, dsn string) *Storage {
	return &Storage{
		dsn:    dsn,
		logger: logger,
	}
}

// Connect establishes database connection with retries.
func (s *Storage) Connect(ctx context.Context) error {
	const maxAttempts = 5
	const retryDelay = 2 * time.Second

	var err error

	for i := 1; i <= maxAttempts; i++ {
		s.db, err = sql.Open("pgx", s.dsn)
		if err != nil {
			err = fmt.Errorf("creating connection pool: %w", err)
		} else if pingErr := s.db.PingContext(ctx); pingErr != nil {
			err = fmt.Errorf("establishing connection: %w", pingErr)
		} else {
			// Успешное подключение
			return nil
		}

		log.Printf("Attempt %d: failed to connect to PostgreSQL: %v", i, err)

		select {
		case <-ctx.Done():
			return fmt.Errorf("connection aborted by context: %w", ctx.Err())
		case <-time.After(retryDelay):
			// Пауза перед следующей попыткой
		}
	}

	return fmt.Errorf("could not connect to PostgreSQL after %d attempts: %w", maxAttempts, err)
}

// Close closes database connection.
func (s *Storage) Close() error {
	return s.db.Close()
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Migrate runs database migrations.
func (s *Storage) Migrate(migrate string) (err error) {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting dialect: %w", err)
	}

	if err := goose.Up(s.db, migrate); err != nil {
		return fmt.Errorf("migration error: %w", err)
	}

	return nil
}

// CreateEvent inserts a new event into database.
func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	ctx = logger.WithLogMethod(ctx, "CreateEvent")
	s.logger.DebugContext(ctx, "attempting to create event")

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
		return logger.WrapError(ctx, err)
	}
	s.logger.InfoContext(ctx, "event created successfully")
	return nil
}

// UpdateEvent updates an existing event in database.
func (s *Storage) UpdateEvent(ctx context.Context, id uuid.UUID, newEvent storage.Event) error {
	ctx = logger.WithLogMethod(ctx, "UpdateEvent")
	s.logger.DebugContext(ctx, "attempting to update event")

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
		return logger.WrapError(ctx, err)
	}
	s.logger.InfoContext(ctx, "event updated successfully")
	return nil
}

// DeleteEvent removes an event from database.
func (s *Storage) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	ctx = logger.WithLogMethod(ctx, "DeleteEvent")

	s.logger.DebugContext(ctx, "attempting to delete event")

	query := `
        DELETE FROM events
        WHERE id = $1
    `

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return logger.WrapError(ctx, err)
	}
	s.logger.InfoContext(ctx, "event deleted successfully")
	return nil
}

// GetEventsDay selects events for one day.
func (s *Storage) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, "Day")
}

// GetEventsWeek selects events for one week.
func (s *Storage) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, "Week")
}

// GetEventsMonth selects events for one month.
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

	s.logger.DebugContext(ctx, "attempting to get events for interval")

	query := `
        SELECT id, title, description, user_id, start_time, end_time, time_before
        FROM events
        WHERE start_time <= $2 AND end_time >= $1
    `

	rows, err := s.db.QueryContext(ctx, query, start, start.Add(d))
	if err != nil {
		return nil, logger.WrapError(ctx, err)
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
			return nil, logger.WrapError(ctx, err)
		}

		dur, err := parsePostgresInterval(intervalStr)
		if err != nil {
			return nil, logger.WrapError(ctx, err)
		}
		event.TimeBefore = dur

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, logger.WrapError(ctx, err)
	}

	s.logger.InfoContext(ctx, "events retrieved successfully", "count", len(events))

	return events, nil
}

// GetNotifications returns upcoming event notifications.
func (s *Storage) GetNotifications(
	ctx context.Context,
	currTime time.Time,
	tick time.Duration,
) ([]storage.Notification, error) {
	ctx = logger.WithLogMethod(ctx, "GetNotifications")
	ctx = logger.WithLogStart(ctx, currTime)

	s.logger.DebugContext(ctx, "attempting to get notifications for interval")

	query := `
        SELECT id, title, start_time, user_id
        FROM events
        WHERE start_time - time_before <= $2 AND start_time - time_before >= $1
    `

	rows, err := s.db.QueryContext(ctx, query, currTime, currTime.Add(tick))
	if err != nil {
		return nil, logger.WrapError(ctx, err)
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
			return nil, logger.WrapError(ctx, err)
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, logger.WrapError(ctx, err)
	}

	s.logger.InfoContext(ctx, "notifications retrieved successfully", "count", len(notifications))

	return notifications, nil
}

// DeleteOldEvents removes events that ended before delTime.
func (s *Storage) DeleteOldEvents(ctx context.Context, delTime time.Time) error {
	ctx = logger.WithLogMethod(ctx, "DeleteOldEvents")
	ctx = logger.WithLogStart(ctx, delTime)

	s.logger.DebugContext(ctx, "attempting to delete old events")

	query := `
        DELETE FROM events
        WHERE end_time < $1
    `

	res, err := s.db.ExecContext(ctx, query, delTime)
	if err != nil {
		return logger.WrapError(ctx, err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return logger.WrapError(ctx, err)
	}
	if count > 0 {
		s.logger.InfoContext(ctx, "old events deleted successfully", "count", count)
	} else {
		s.logger.InfoContext(ctx, "no old events to delete")
	}
	return nil
}
