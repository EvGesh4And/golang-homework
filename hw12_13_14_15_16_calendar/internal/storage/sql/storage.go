package sqlstorage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"time"

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
		return err
	}

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, id uuid.UUID, newEvent storage.Event) error {
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
		return err
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	query := `
        DELETE FROM events
        WHERE id = $1
    `

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetEventsDay(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, time.Hour*24)
}

func (s *Storage) GetEventsWeek(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, time.Hour*24*7)
}

func (s *Storage) GetEventsMonth(ctx context.Context, start time.Time) ([]storage.Event, error) {
	return s.getEvents(ctx, start, time.Hour*24*30)
}

func (s *Storage) getEvents(ctx context.Context, start time.Time, d time.Duration) ([]storage.Event, error) {
	end := start.Add(d)

	query := `
        SELECT id, title, description, user_id, start_time, end_time, time_before
        FROM events
        WHERE start_time <= $2 AND end_time >= $1
    `

	rows, err := s.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		dur, err := parsePostgresInterval(intervalStr)
		if err != nil {
			return nil, err
		}
		event.TimeBefore = dur

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
