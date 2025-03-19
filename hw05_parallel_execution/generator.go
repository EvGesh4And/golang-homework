package hw05parallelexecution

import "sync"

// generator создает канал задач и передает их, пока не получит сигнал завершения.
func generator(tasks []Task, stopCh <-chan struct{}, wg *sync.WaitGroup) <-chan Task {
	taskStream := make(chan Task)
	go func() {
		defer wg.Done()
		defer close(taskStream)
		for _, task := range tasks {
			select {
			case <-stopCh:
				return
			default:
				select {
				case <-stopCh:
					return
				case taskStream <- task:
				}
			}
		}
	}()
	return taskStream
}
