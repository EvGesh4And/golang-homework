package hw05parallelexecution

import "sync"

// worker выполняет задачи из taskStream и завершает работу по сигналу в done.
func worker(taskStream <-chan Task, stopper *Stopper, wg *sync.WaitGroup) {
	defer wg.Done() // Уменьшаем счётчик в wg по завершению работы воркера

	for task := range taskStream {
		if err := task(); err != nil { // Выполняем задачу и обрабатываем ошибку
			stopper.AddError()
		}
	}
}
