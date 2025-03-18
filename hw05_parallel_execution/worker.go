package hw05parallelexecution

import "sync"

// worker выполняет задачи из taskStream и завершает работу по сигналу в done.
func worker(done <-chan struct{}, taskStream <-chan Task, stopper *Stopper, wg *sync.WaitGroup) {
	defer wg.Done() // Уменьшаем счётчик в wg по завершению работы воркера

	for {
		select {
		case <-done: // Завершаем работу по сигналу из done
			return
		default:
			select {
			case <-done: // Завершаем работу по сигналу из done
				return
			case task, ok := <-taskStream: // Получаем задачу из канала
				if !ok { // Если канал закрыт, отправляем сигнал остановки
					stopper.SignalStop()
					return
				}
				if err := task(); err != nil { // Выполняем задачу и обрабатываем ошибку
					stopper.AddError()
				}
			}
		}
	}
}
