package memorystorage

import (
	"testing"
	"time"

	"github.com/EvGesh4And/hw12_13_14_15_calendar/internal/logger"
	"github.com/EvGesh4And/hw12_13_14_15_calendar/internal/storage"
)

func createTestEvent(id string, title string, start time.Time, duration time.Duration) storage.Event {
	return storage.Event{
		IDEvent: storage.IDEvent(id),
		Title:   title,
		Start:   start,
		End:     start.Add(duration),
	}
}

func TestStorage_AddUpdateDelete(t *testing.T) {
	logg := logger.New("debug")
	store := New(logg)

	start := time.Now().Add(time.Minute)
	event := createTestEvent("1", "Test Event", start, time.Hour)

	// Добавление
	err := store.AddEvent(event)
	if err != nil {
		t.Fatalf("failed to add event: %v", err)
	}

	// Повторное добавление того же ID — должно вернуть ошибку
	err = store.AddEvent(event)
	if err != storage.ErrIDEventRepeated {
		t.Errorf("expected ErrIDEventRepeated, got: %v", err)
	}

	// Обновление
	newEvent := createTestEvent("1", "Updated Event", start.Add(time.Hour*2), time.Hour)
	err = store.UpdateEvent(event.IDEvent, newEvent)
	if err != nil {
		t.Errorf("failed to update event: %v", err)
	}

	// Удаление
	err = store.DeleteEvent(event.IDEvent)
	if err != nil {
		t.Errorf("failed to delete event: %v", err)
	}

	// Повторное удаление — ошибка
	err = store.DeleteEvent(event.IDEvent)
	if err != storage.ErrIDEventNotExist {
		t.Errorf("expected ErrIDEventNotExist, got: %v", err)
	}
}

func TestStorage_GetEvents(t *testing.T) {
	logg := logger.New("debug")
	store := New(logg)

	now := time.Now().Add(time.Minute)

	// Добавим события в разные дни
	events := []storage.Event{
		createTestEvent("1", "Today", now, time.Hour),
		createTestEvent("2", "Tomorrow", now.Add(25*time.Hour), time.Hour),
		createTestEvent("3", "In Week", now.Add(6*24*time.Hour), time.Hour),
	}

	for _, e := range events {
		if err := store.AddEvent(e); err != nil {
			t.Fatalf("failed to add event %s: %v", e.IDEvent, err)
		}
	}

	dayEvents, err := store.GetEventsDay(now)
	if err != nil {
		t.Fatalf("GetEventsDay failed: %v", err)
	}
	if len(dayEvents) != 1 {
		t.Errorf("expected 1 event today, got %d", len(dayEvents))
	}

	weekEvents, err := store.GetEventsWeek(now)
	if err != nil {
		t.Fatalf("GetEventsWeek failed: %v", err)
	}
	if len(weekEvents) != 3 {
		t.Errorf("expected 3 events this week, got %d", len(weekEvents))
	}
}
