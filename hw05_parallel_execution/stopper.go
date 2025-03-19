package hw05parallelexecution

import "sync"

// Stopper отслеживает количество ошибок и отправляет сигнал остановки,
// если количество ошибок достигает установленного лимита.
type Stopper struct {
	stopCh          chan<- struct{} // Канал для отправки сигнала остановки
	m               sync.Mutex      // Мьютекс для синхронизации доступа к данным
	count           int             // Счётчик ошибок
	limit           int             // Лимит ошибок, при достижении которого будет отправлен сигнал остановки
	statusOverLimit bool            // Флаг, указывающий, что лимит ошибок был достигнут
}

// NewErrorStopper создаёт новый объект ErrorStopper с заданным каналом и лимитом ошибок.
func NewStopper(stopCh chan struct{}, limit int) *Stopper {
	return &Stopper{
		stopCh: stopCh,
		limit:  limit,
	}
}

// AddError увеличивает счётчик ошибок и проверяет, достигнут ли лимит ошибок.
// Если лимит достигнут, вызывается метод SignalStop для отправки сигнала об остановке.
func (s *Stopper) AddError() {
	s.m.Lock()
	defer s.m.Unlock()

	s.count++

	// Если лимит ошибок ещё не был достигнут, проверяем, достигнут ли лимит.
	if !s.statusOverLimit && s.count >= s.limit {
		s.SignalStop()
		s.statusOverLimit = true
	}
}

// SignalStop отправляет сигнал в канал stop, уведомляя другие горутины или процессы
// о необходимости остановки.
func (s *Stopper) SignalStop() {
	s.stopCh <- struct{}{}
}
