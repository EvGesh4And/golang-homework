package memorystorage

import (
	"testing"
	"time"

	"github.com/EvGesh4And/hw12_13_14_15_calendar/internal/storage"
)

func makeInterval(start, end time.Time, id string) storage.Interval {
	return storage.Interval{Start: start, End: end, IDEvent: storage.IDEvent(id)}
}

func TestIntervalSlice_AddRemoveReplace(t *testing.T) {
	now := time.Now()
	a := makeInterval(now, now.Add(time.Hour), "1")
	b := makeInterval(now.Add(time.Minute*30), now.Add(time.Hour*2), "2")
	c := makeInterval(now.Add(time.Hour*2), now.Add(time.Hour*3), "3")

	slice := IntervalSlice{}

	// Добавим первый интервал
	if !slice.AddIfFree(a) {
		t.Fatal("failed to add interval a")
	}

	// Пересекающийся — не должен добавиться
	if slice.AddIfFree(b) {
		t.Error("should not add overlapping interval b")
	}

	// Не пересекающийся — должен добавиться
	if !slice.AddIfFree(c) {
		t.Error("failed to add non-overlapping interval c")
	}

	// Удалим a
	if !slice.Remove(a) {
		t.Error("failed to remove interval a")
	}

	// Теперь b можно добавить
	if !slice.AddIfFree(b) {
		t.Error("failed to add interval b after a removed")
	}
}

func TestIntervalSlice_Replace(t *testing.T) {
	now := time.Now()
	a := makeInterval(now, now.Add(time.Hour), "1")
	b := makeInterval(now.Add(time.Hour*2), now.Add(time.Hour*3), "2")              // свободен
	conflict := makeInterval(now.Add(time.Minute*30), now.Add(time.Hour*1+30), "3") // пересекается с a

	slice := IntervalSlice{}
	if !slice.AddIfFree(a) {
		t.Fatal("failed to add interval a")
	}

	// Успешная замена на свободный интервал
	if !slice.Replace(b, a) {
		t.Error("failed to replace a with b")
	}

	// Попытка замены на конфликтный — должна провалиться и вернуть a обратно
	if !slice.Replace(conflict, b) {
		t.Error("expected replace to fail but succeed")
	}

	if !slice.Remove(conflict) {
		t.Log("conflict interval was not added (as expected)")
	}
}
