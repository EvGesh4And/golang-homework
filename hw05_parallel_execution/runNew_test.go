package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
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
				for range 1_000_000 {
					s += 5 * 5 / 5
				}

				time.Sleep(time.Millisecond)

				for range 1_000_000 {
					s += 5 * 5 / 5
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

func TestRunLastError(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("error in the last task, with a one-error limit", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		// Задачи без ошибок
		for range tasksCount - 1 {
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}
		// Последняя задача с ошибкой
		tasks = append(tasks, func() error {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			atomic.AddInt32(&runTasksCount, 1)
			return fmt.Errorf("error from task %d", 49)
		})

		workersCount := 10
		maxErrorsCount := 1
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(tasksCount), "extra tasks were started")
	})
}
