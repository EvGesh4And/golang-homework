package memorystorage

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
)

// Вспомогательная функция для создания тестового события.
func createTestEvent(id uuid.UUID, title string, start time.Time, duration time.Duration) storage.Event {
	return storage.Event{
		ID:    id,
		Title: title,
		Start: start,
		End:   start.Add(duration),
	}
}

// Тест: добавление, обновление и удаление события.
func TestStorage_AddUpdateDelete(t *testing.T) {
	ctx := context.Background()

	logger := logger.New("info", os.Stdout)
	store := New(logger)

	start := time.Now().Add(time.Minute)
	event := createTestEvent(uuid.New(), "Тестовое событие", start, time.Hour)

	// Добавление.
	err := store.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("не удалось добавить событие: %v", err)
	}

	// Повторное добавление того же ID — должно вернуть ошибку.
	err = store.CreateEvent(ctx, event)
	if !errors.Is(err, storage.ErrIDRepeated) {
		t.Errorf("ожидалась ошибка ErrIDRepeated, получено: %v", err)
	}

	// Обновление.
	newEvent := createTestEvent(uuid.New(), "Обновлённое событие", start.Add(time.Hour*2), time.Hour)
	err = store.UpdateEvent(ctx, event.ID, newEvent)
	if err != nil {
		t.Errorf("не удалось обновить событие: %v", err)
	}

	// Удаление.
	err = store.DeleteEvent(ctx, event.ID)
	if err != nil {
		t.Errorf("не удалось удалить событие: %v", err)
	}

	// Повторное удаление — ожидается ошибка.
	err = store.DeleteEvent(ctx, event.ID)
	if !errors.Is(err, storage.ErrIDNotExist) {
		t.Errorf("ожидалась ошибка ErrIDNotExist, получено: %v", err)
	}
}

// Тест: получение событий за день и неделю.
func TestStorage_GetEvents(t *testing.T) {
	ctx := context.Background()

	logger := logger.New("info", os.Stdout)
	store := New(logger)

	now := time.Now().Add(time.Minute)

	// Добавим события в разные дни.
	events := []storage.Event{
		createTestEvent(uuid.New(), "Сегодня", now, time.Hour),
		createTestEvent(uuid.New(), "Завтра", now.Add(25*time.Hour), time.Hour),
		createTestEvent(uuid.New(), "Через неделю", now.Add(6*24*time.Hour), time.Hour),
	}

	for _, e := range events {
		if err := store.CreateEvent(ctx, e); err != nil {
			t.Fatalf("не удалось добавить событие %s: %v", e.ID, err)
		}
	}

	dayEvents, err := store.GetEventsDay(ctx, now)
	if err != nil {
		t.Fatalf("ошибка при получении событий за день: %v", err)
	}
	if len(dayEvents) != 1 {
		t.Errorf("ожидалось 1 событие на сегодня, получено: %d", len(dayEvents))
	}

	weekEvents, err := store.GetEventsWeek(ctx, now)
	if err != nil {
		t.Fatalf("ошибка при получении событий за неделю: %v", err)
	}
	if len(weekEvents) != 3 {
		t.Errorf("ожидалось 3 события на этой неделе, получено: %d", len(weekEvents))
	}
}

// Тест: потокобезопасность при параллельном доступе.
func TestStorage_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()

	logger := logger.New("info", os.Stdout)
	store := New(logger)

	start := time.Now().Add(time.Minute)

	const goroutines = 100
	errCh := make(chan error, goroutines*2)

	// Параллельно добавляем события.
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			id := uuid.New()
			event := createTestEvent(id, "Событие", start.Add(time.Duration(i)*time.Minute), time.Second)
			err := store.CreateEvent(ctx, event)
			errCh <- err
		}(i)
	}

	// Параллельно удаляем события.
	for i := 0; i < goroutines; i++ {
		go func() {
			id := uuid.New()
			err := store.DeleteEvent(ctx, id)
			// Ошибка может быть нормальной, если удаление происходит до добавления.
			if err != nil && !errors.Is(err, storage.ErrIDNotExist) {
				errCh <- err
			} else {
				errCh <- nil
			}
		}()
	}

	// Проверим ошибки.
	for i := 0; i < goroutines*2; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("обнаружена ошибка при параллельном доступе: %v", err)
		}
	}
}
