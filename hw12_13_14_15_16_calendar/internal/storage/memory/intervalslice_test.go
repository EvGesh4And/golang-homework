package memorystorage

import (
	"testing"
	"time"

	"github.com/EvGesh4And/hw12_13_14_15_calendar/internal/storage"
)

// Вспомогательная функция для создания интервала.
func makeInterval(start, end time.Time, id string) storage.Interval {
	return storage.Interval{Start: start, End: end, ID: id}
}

// Тест: добавление, удаление и замена интервалов.
func TestIntervalSlice_AddRemoveReplace(t *testing.T) {
	now := time.Now()
	a := makeInterval(now, now.Add(time.Hour), "1")
	b := makeInterval(now.Add(time.Minute*30), now.Add(time.Hour*2), "2")
	c := makeInterval(now.Add(time.Hour*2), now.Add(time.Hour*3), "3")

	slice := IntervalSlice{}

	// Добавим первый интервал.
	if !slice.AddIfFree(a) {
		t.Fatal("не удалось добавить интервал a")
	}

	// Пересекающийся интервал — не должен добавиться.
	if slice.AddIfFree(b) {
		t.Error("не должен был добавиться пересекающийся интервал b")
	}

	// Не пересекающийся интервал — должен добавиться.
	if !slice.AddIfFree(c) {
		t.Error("не удалось добавить непересекающийся интервал c")
	}

	// Удалим a.
	if !slice.Remove(a) {
		t.Error("не удалось удалить интервал a")
	}

	// Теперь b можно добавить.
	if !slice.AddIfFree(b) {
		t.Error("не удалось добавить интервал b после удаления a")
	}
}

// Тест: замена интервалов.
func TestIntervalSlice_Replace(t *testing.T) {
	now := time.Now()
	a := makeInterval(now, now.Add(time.Hour), "1")
	b := makeInterval(now.Add(time.Hour*2), now.Add(time.Hour*3), "2")              // свободный интервал.
	conflict := makeInterval(now.Add(time.Minute*30), now.Add(time.Hour*1+30), "3") // пересекается с a.

	slice := IntervalSlice{}
	if !slice.AddIfFree(a) {
		t.Fatal("не удалось добавить интервал a")
	}

	// Успешная замена на свободный интервал.
	if !slice.Replace(b, a) {
		t.Error("не удалось заменить a на b")
	}

	// Попытка замены на пересекающийся интервал — должна завершиться неудачей, b должен быть возвращён.
	if !slice.Replace(conflict, b) {
		t.Error("ожидалась ошибка замены, но замена прошла успешно")
	}

	if !slice.Remove(conflict) {
		t.Log("пересекающийся интервал не был добавлен (как и ожидалось)")
	}
}
