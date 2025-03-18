package hw05parallelexecution

import "sync"

// generator создает канал задач и передает их, пока не получит сигнал завершения.
func generator(done <-chan struct{}, tasks []Task, n int, wg *sync.WaitGroup) <-chan Task {
	defer wg.Done()
	taskStream := make(chan Task, n)
	go func() {
		defer close(taskStream)
		for _, task := range tasks {
			select {
			case <-done:
				return
			default:
				select {
				case <-done:
					return
				case taskStream <- task:
				}
			}
		}
	}()
	return taskStream
}
