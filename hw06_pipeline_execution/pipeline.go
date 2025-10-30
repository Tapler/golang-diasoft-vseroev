package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	// Если стейджей нет — всё равно оборачиваем входной канал обёрткой, чтобы потребитель мог прерваться по done.
	if len(stages) == 0 {
		return doneAwareChan(in, done)
	}
	current := in
	for _, stage := range stages {
		current = doneAwareChan(current, done) // Оборачиваем текущий канал в done-aware обертку
		current = stage(current)
	}
	return doneAwareChan(current, done)
}

// doneAwareChan — обертка, реализующая паттерн for-select с <-done.
func doneAwareChan(in In, done In) Out {
	// если входной канал nil — считаем его закрытым и сразу закрываем выход.
	if in == nil {
		ch := make(Bi)
		close(ch)
		return ch
	}

	// Создаём выходной канал для передачи данных
	out := make(Bi)

	// Мониторим done на чтении и записи.
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case val, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-done:
					return
				case out <- val:
				}
			}
		}
	}()

	return out
}
