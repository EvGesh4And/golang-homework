package storage

import "errors"

var (
	// ошибка в полях события
	ErrEventValidStart  = errors.New("ошибка в дате и времени начала события")
	ErrEventValidEnd    = errors.New("ошибка в длительности события")
	ErrEventValidBefore = errors.New("ошибка в времени заблаговременного уведомления")
	// ошибка с ID
	ErrIDEventRepeated = errors.New("событие с таким ID уже есть в хранилище")
	ErrIDEventNotExist = errors.New("события с таким ID не существует")
	// ошибка с временными интервалами
	ErrDateBusy = errors.New("данное время уже занято другим событием")
	// ошибка с получем событий в интервале
	ErrGetEvents = errors.New("ошибка в ходе получения списка событий")
)
