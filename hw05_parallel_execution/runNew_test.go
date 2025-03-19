package hw05parallelexecution

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRunEventually(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var activeWorkers int32
		workersCount := 5
		maxErrorsCount := 1

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				atomic.AddInt32(&activeWorkers, 1)        // Увеличиваем счетчик активных воркеров
				defer atomic.AddInt32(&activeWorkers, -1) // Уменьшаем после завершения

				s := 0
				for i := 0; i < 1_000_000; i++ {
					s += i * (i + 1) / (i + 1)
				}

				time.Sleep(time.Millisecond)

				for i := 0; i < 1_000_000; i++ {
					s += i * (i + 1) / (i + 1)
				}

				atomic.AddInt32(&runTasksCount, 1) // Увеличиваем счетчик выполненных задач
				return nil
			})
		}

		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			err := Run(tasks, workersCount, maxErrorsCount)
			require.NoError(t, err)
		}()

		require.Eventually(t, func() bool {
			return atomic.LoadInt32(&activeWorkers) > 1
		}, 2*time.Second, 10*time.Millisecond)

		wg.Wait()
		// Проверяем, что все задачи выполнены
		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
	})
}
