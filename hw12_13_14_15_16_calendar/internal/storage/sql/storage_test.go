package sqlstorage

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
)

var testDSN = os.Getenv("TEST_DSN")

func setupStorage(t *testing.T) *Storage {
	t.Helper()

	st := New(&slog.Logger{}, testDSN)

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
