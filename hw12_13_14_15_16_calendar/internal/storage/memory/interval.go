package memorystorage

import "github.com/EvGesh4And/hw12_13_14_15_calendar/internal/storage"

type IntervalSlice struct {
	Intervals []storage.Interval
}

// Проверяет, можно ли добавить интервал без пересечений
func (s *IntervalSlice) CanAdd(newInterval storage.Interval) bool {
	for _, interval := range s.Intervals {
		if intervalsOverlap(interval, newInterval) {
			return false
		}
	}
	return true
}

// Добавляет интервал, если нет пересечений. Возвращает true, если добавлен.
func (s *IntervalSlice) AddIfFree(newInterval storage.Interval) bool {
	if s.CanAdd(newInterval) {
		s.Intervals = append(s.Intervals, newInterval)
		return true
	}
	return false
}

// Удаляет точное совпадение интервала. Возвращает true, если удалён.
func (s *IntervalSlice) Remove(target storage.Interval) bool {
	for i, interval := range s.Intervals {
		if interval.Start.Equal(target.Start) && interval.End.Equal(target.End) {
			s.Intervals = append(s.Intervals[:i], s.Intervals[i+1:]...)
			return true
		}
	}
	return false
}

func (s *IntervalSlice) Replace(newInterval, oldInterval storage.Interval) bool {
	if !s.Remove(oldInterval) {
		return false
	}
	if s.CanAdd(newInterval) {
		s.Intervals = append(s.Intervals, newInterval)
		return true
	}
	// Откат
	s.Intervals = append(s.Intervals, oldInterval)
	return false
}

func (s *IntervalSlice) GetStartedInInterval(interval storage.Interval) []storage.Interval {
	var res []storage.Interval

	for _, inter := range s.Intervals {
		if startInInterval(inter, interval) {
			res = append(res, inter)
		}
	}
	return res
}

// Проверяет пересечение интервалов
func intervalsOverlap(a, b storage.Interval) bool {
	return a.Start.Before(b.End) && b.Start.Before(a.End)
}

// Проверяет пересечение интервалов
func startInInterval(a, b storage.Interval) bool {
	return (b.Start.Equal(a.Start) || b.Start.Before(a.Start)) && (a.Start.Before(b.End) || a.Start.Equal(b.End))
}
