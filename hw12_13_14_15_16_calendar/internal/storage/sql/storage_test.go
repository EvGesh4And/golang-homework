//go:build integration
// +build integration

package sqlstorage_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	sqlstorage "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage/sql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupStorage(t *testing.T) *sqlstorage.Storage {

	if testing.Short() {
		t.Skip("-short: пропускаем интеграцию")
	}

	t.Helper()

	logger := logger.New("info", os.Stdout)

	ctx := context.Background()
	pg, err := testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "postgres:16-alpine",
				Env: map[string]string{
					"POSTGRES_USER":     "otus_user",
					"POSTGRES_PASSWORD": "otus_password",
					"POSTGRES_DB":       "otus",
				},
				ExposedPorts: []string{"5432/tcp"},
				WaitingFor:   wait.ForListeningPort("5432/tcp"),
			},
			Started: true,
		})
	require.NoError(t, err)
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	host, _ := pg.Host(ctx)
	port, _ := pg.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("host=%s port=%s user=otus_user password=otus_password dbname=otus sslmode=disable",
		host, port.Port())
	fmt.Println(dsn)
	st := sqlstorage.New(logger, dsn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := st.Connect(ctx); err != nil {
		t.Fatalf("ошибка подключения: %v", err)
	}

	if err := st.Migrate("migrations"); err != nil {
		t.Fatalf("ошибка миграции: %v", err)
	}

	t.Cleanup(func() {
		_ = st.Close()
	})

	return st
}

func makeTestEvent() storage.Event {
	return storage.Event{
		ID:          uuid.New(),
		Title:       "Test Event",
		Description: "Test Description",
		UserID:      uuid.New(),
		Start:       time.Now().Add(time.Hour),
		End:         time.Now().Add(3 * time.Hour),
		TimeBefore:  15 * time.Minute,
	}
}

func TestCreateAndGetEvent(t *testing.T) {
	st := setupStorage(t)

	ctx := context.Background()
	event := makeTestEvent()

	err := st.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}

	events, err := st.GetEventsDay(ctx, event.Start)
	if err != nil {
		t.Fatalf("GetEventsDay: %v", err)
	}

	found := false
	for _, e := range events {
		if e.ID == event.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("созданное событие не найдено в результате GetEventsDay")
	}
}

func TestUpdateEvent(t *testing.T) {
	st := setupStorage(t)

	ctx := context.Background()
	event := makeTestEvent()
	_ = st.CreateEvent(ctx, event)

	newTitle := "Updated Title"
	event.Title = newTitle

	err := st.UpdateEvent(ctx, event.ID, event)
	if err != nil {
		t.Fatalf("UpdateEvent: %v", err)
	}

	events, _ := st.GetEventsDay(ctx, event.Start)
	for _, e := range events {
		if e.ID == event.ID && e.Title != newTitle {
			t.Errorf("ожидаемый обновленный заголовок %q, полученный %q", newTitle, e.Title)
		}
	}
}

func TestDeleteEvent(t *testing.T) {
	st := setupStorage(t)

	ctx := context.Background()
	event := makeTestEvent()
	_ = st.CreateEvent(ctx, event)

	err := st.DeleteEvent(ctx, event.ID)
	if err != nil {
		t.Fatalf("DeleteEvent: %v", err)
	}

	events, _ := st.GetEventsDay(ctx, event.Start)
	for _, e := range events {
		if e.ID == event.ID {
			t.Errorf("событие не удалено")
		}
	}
}

func TestDeleteOldEvents(t *testing.T) {
	st := setupStorage(t)

	ctx := context.Background()
	now := time.Now()

	oldEvent := storage.Event{
		ID:          uuid.New(),
		Title:       "Old Event",
		Description: "old",
		UserID:      uuid.New(),
		Start:       now.Add(-2 * time.Hour),
		End:         now.Add(-1 * time.Hour),
		TimeBefore:  15 * time.Minute,
	}
	newEvent := storage.Event{
		ID:          uuid.New(),
		Title:       "New Event",
		Description: "new",
		UserID:      uuid.New(),
		Start:       now.Add(time.Hour),
		End:         now.Add(2 * time.Hour),
		TimeBefore:  15 * time.Minute,
	}

	if err := st.CreateEvent(ctx, oldEvent); err != nil {
		t.Fatalf("CreateEvent old: %v", err)
	}
	if err := st.CreateEvent(ctx, newEvent); err != nil {
		t.Fatalf("CreateEvent new: %v", err)
	}

	if err := st.DeleteOldEvents(ctx, now); err != nil {
		t.Fatalf("DeleteOldEvents: %v", err)
	}

	events, err := st.GetEventsMonth(ctx, now.Add(-3*time.Hour))
	if err != nil {
		t.Fatalf("GetEventsMonth: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event after deletion, got %d", len(events))
	}
	if events[0].ID != newEvent.ID {
		t.Errorf("unexpected event remaining: %v", events[0].ID)
	}
}
