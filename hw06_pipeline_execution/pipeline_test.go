package hw06pipelineexecution

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

func TestPipeline(t *testing.T) {
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})

	t.Run("no stages", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]int, 0)
		for val := range ExecutePipeline(in, nil) {
			result = append(result, val.(int))
		}

		// Должны получить все исходные данные без изменений
		require.Equal(t, data, result)
	})

	t.Run("done before data", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)

		// Закрываем done сразу
		close(done)

		go func() {
			time.Sleep(sleepPerStage)
			for i := 1; i <= 5; i++ {
				in <- i
			}
			close(in)
		}()

		result := make([]string, 0)
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}

		// Не должно быть обработано ни одного элемента
		require.Len(t, result, 0)
	})

	t.Run("empty input", func(t *testing.T) {
		in := make(Bi)
		close(in) // Сразу закрываем канал

		result := make([]string, 0)
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}

		// Результат должен быть пустым
		require.Len(t, result, 0)
	})
}
