package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

var ErrErrorsLimitWorkers = errors.New("number of workers must be greater than zero")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		return ErrErrorsLimitWorkers
	}
	tasksCh := make(chan Task)
	// WaitGroup для ожидания завершения всех воркеров
	wg := sync.WaitGroup{}

	// игнорируем ошибки если m <= 0
	ignoreErrors := m <= 0
	var errCount int
	var mu sync.Mutex             // мьютекс для защиты errCount
	stopCh := make(chan struct{}) // канал для досрочного завершения воркеров
	var stopOnce sync.Once

	// буферизованный канал для сбора ошибок от воркеров
	// размер буфера = количество воркеров, чтобы избежать блокировок
	errCh := make(chan error, n)

	// воркер выполняет задачи из tasksCh
	worker := func() {
		defer wg.Done() // уменьшает счетчик на 1, когда воркер завершает работу
		for {
			select {
			case <-stopCh:
				// Получен сигнал остановки — выходим
				return
			case task, ok := <-tasksCh:
				if !ok {
					// Задачи закончились — выходим
					return
				}
				errCh <- task() // Выполняем задачу и отправляем ошибку или nil
			}
		}
	}

	// Запускаем n воркеров
	for i := 0; i < n; i++ {
		wg.Add(1) // увеличиваем счетчик WaitGroup
		go worker()
	}

	// Горутина для отправки задач в канал
	go func() {
		for _, task := range tasks {
			select {
			case <-stopCh:
				// Если получен сигнал остановки — прекращаем отправку задач
				return
			case tasksCh <- task:
			}
		}
		close(tasksCh) // Все задачи отправлены — закрываем канал
	}()

	// Горутина для закрытия errCh после завершения всех воркеров
	go func() {
		wg.Wait()    // ждём завершения всех воркеров
		close(errCh) // закрываем канал ошибок
	}()

	// собираем ошибки от воркеров в цикле
	// цикл завершится когда errCh будет закрыт и все ошибки обработаны
	for err := range errCh {
		if err != nil && !ignoreErrors {
			mu.Lock()
			errCount++
			// Досрочное завершение если достигли лимита m
			if errCount >= m {
				stopOnce.Do(func() { close(stopCh) })
			}
			mu.Unlock()
		}
	}

	// Проверяем превышен ли лимит ошибок с защитой от race condition
	mu.Lock()
	shouldReturnError := !ignoreErrors && errCount >= m
	mu.Unlock()

	if shouldReturnError {
		return ErrErrorsLimitExceeded
	}
	return nil
}
