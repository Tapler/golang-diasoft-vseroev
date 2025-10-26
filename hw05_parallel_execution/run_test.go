package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})

	t.Run("m <= 0: all tasks run, errors ignored", func(t *testing.T) {
		tasksCount := 20
		tasks := make([]Task, 0, tasksCount)
		var runTasksCount int32
		var errorsCount int32

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				atomic.AddInt32(&runTasksCount, 1)
				err := fmt.Errorf("fail")
				atomic.AddInt32(&errorsCount, 1)
				return err
			})
		}

		workersCount := 4
		maxErrorsCount := 0 // m <= 0 — ошибки игнорируются
		err := Run(tasks, workersCount, maxErrorsCount)

		// Ошибка не должна возвращаться
		require.NoError(t, err)
		// Все задачи должны быть выполнены
		require.Equal(t, int32(tasksCount), runTasksCount)
		// Все задачи вернули ошибку, но ошибки проигнорированы
		require.Equal(t, int32(tasksCount), errorsCount)
	})

	t.Run("tasks less than workers", func(t *testing.T) {
		tasksCount := 3
		tasks := make([]Task, 0, tasksCount)
		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 10
		maxErrorsCount := 1
		err := Run(tasks, workersCount, maxErrorsCount)

		// Ошибок нет и все задачи выполнены
		require.NoError(t, err)
		require.Equal(t, int32(tasksCount), runTasksCount)
	})

	t.Run("no tasks", func(t *testing.T) {
		tasks := []Task{}
		workersCount := 5
		maxErrorsCount := 2
		err := Run(tasks, workersCount, maxErrorsCount)
		require.NoError(t, err)
	})

	t.Run("n = 0 should return error", func(t *testing.T) {
		tasks := []Task{func() error { return nil }}
		err := Run(tasks, 0, 1)
		require.EqualError(t, err, "number of workers must be greater than zero")
	})

	t.Run("no goroutine leaks", func(t *testing.T) {
		defer goleak.VerifyNone(t)
		tasks := []Task{}
		for i := 0; i < 10; i++ {
			tasks = append(tasks, func() error { return nil })
		}
		err := Run(tasks, 3, 2)
		require.NoError(t, err)
	})
}
