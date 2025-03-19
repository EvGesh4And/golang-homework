package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	//Канал, являющийся сигналом для генератора - необходимо остановиться (буфер n, чтобы не было блокировок)
	stopCh := make(chan struct{}, n)

	// Создаём объект для управления остановкой со слежением за лимитом ошибок
	stopper := NewStopper(stopCh, m)

	// Ожидание завершения генератора и всех воркеров
	var wg sync.WaitGroup

	// Генерация задач и отправка их в канал taskStream
	wg.Add(1)
	taskStream := generator(tasks, stopCh, &wg)

	// Запуск n воркеров
	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker(taskStream, stopper, &wg)
	}

	wg.Wait()

	// Если лимит ошибок превышен, возвращаем ошибку
	if stopper.statusOverLimit {
		return ErrErrorsLimitExceeded
	}
	// Если ошибок не было, возвращаем nil
	return nil
}
