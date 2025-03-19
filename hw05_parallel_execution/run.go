package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	var wg sync.WaitGroup
	errorCount := atomic.Int64{}
	taskStream := make(chan Task)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(taskStream)
		for _, task := range tasks {
			if errorCount.Load() >= int64(m) {
				return
			}
			taskStream <- task
		}
	}()

	// Запуск n воркеров
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskStream {
				if err := task(); err != nil {
					errorCount.Add(1)
				}
			}
		}()
	}

	wg.Wait()

	// Если лимит ошибок превышен, возвращаем ошибку
	if errorCount.Load() >= int64(m) {
		return ErrErrorsLimitExceeded
	}
	// Если ошибок не было, возвращаем nil
	return nil
}
