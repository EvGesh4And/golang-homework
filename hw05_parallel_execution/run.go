package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	// Канал для отправки сигнала завершения работы (для генератора и воркеров)
	done := make(chan struct{})
	// Канал для остановки всех горутин (буфер n, чтобы не было блокировок)
	stop := make(chan struct{}, n)

	// Создаём объект для управления остановкой со слежением за лимитом ошибок
	stopper := NewStopper(stop, m)

	// Ожидание завершения генератора и всех воркеров
	var wg sync.WaitGroup

	// Горутину для отправки сигнала завершения работы после получения сигнала из stop
	go func() {
		defer close(done) // Закрываем канал done по завершению работы
		<-stop            // Ожидаем получения первого сигнала из stop
	}()

	// Генерация задач и отправка их в канал taskStream
	wg.Add(1)
	taskStream := generator(done, tasks, n, &wg)

	// Запуск n воркеров
	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker(done, taskStream, stopper, &wg)
	}

	// Ожидаем завершения всех воркеров
	wg.Wait()

	// Если лимит ошибок превышен, возвращаем ошибку
	if stopper.statusOverLimit {
		return ErrErrorsLimitExceeded
	}
	// Если ошибок не было, возвращаем nil
	return nil
}
