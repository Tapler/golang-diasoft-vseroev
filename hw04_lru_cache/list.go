package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	front *ListItem // первый элемент списка (голова)
	back  *ListItem // последний элемент списка (хвост)
	len   int       // текущая длина списка
}

// NewList создаёт новый пустой двусвязный список.
func NewList() List {
	return new(list)
}

// Len возвращает длину списка.
func (l *list) Len() int {
	return l.len
}

// Front возвращает первый элемент списка.
func (l *list) Front() *ListItem {
	return l.front
}

// Back возвращает последний элемент списка.
func (l *list) Back() *ListItem {
	return l.back
}

// PushFront добавляет значение в начало списка.
func (l *list) PushFront(v interface{}) *ListItem {
	// Создаём новый элемент, который будет первым
	item := &ListItem{
		Value: v,
		Next:  l.front, // новый элемент указывает на текущий первый
		Prev:  nil,     // у первого элемента нет предыдущего
	}

	if l.front != nil {
		// Если список не пустой, обновляем ссылку у старого первого элемента
		l.front.Prev = item
	} else {
		// Если список был пустой, новый элемент также становится последним
		l.back = item
	}

	// Обновляем указатель на первый элемент
	l.front = item
	l.len++

	return item
}

// PushBack добавляет значение в конец списка.
func (l *list) PushBack(v interface{}) *ListItem {
	// Создаём новый элемент, который будет последним
	item := &ListItem{
		Value: v,
		Next:  nil,    // у последнего элемента нет следующего
		Prev:  l.back, // новый элемент указывает на текущий последний
	}

	if l.back != nil {
		// Если список не пустой, обновляем ссылку у прошлого последнего элемента
		l.back.Next = item
	} else {
		// Если список был пустой, новый элемент также становится первым
		l.front = item
	}

	// Обновляем указатель на последний элемент
	l.back = item
	l.len++

	return item
}

// Remove удаляет элемент из списка.
// Предполагается, что элемент обзательно существует в списке.
func (l *list) Remove(i *ListItem) {
	// Обновляем ссылку у предыдущего элемента
	if i.Prev != nil {
		// Если есть предыдущий элемент, связываем его со следующим
		i.Prev.Next = i.Next
	} else {
		// Если удаляем первый элемент, обновляем указатель front
		l.front = i.Next
	}

	// Обновляем ссылку у следующего элемента
	if i.Next != nil {
		// Если есть следующий элемент, связываем его с предыдущим
		i.Next.Prev = i.Prev
	} else {
		// Если удаляем последний элемент, обновляем указатель back
		l.back = i.Prev
	}

	// Уменьшаем длину списка
	l.len--
}

// MoveToFront перемещает элемент в начало списка.
// Предполагается, что элемент существует в списке.
func (l *list) MoveToFront(i *ListItem) {
	// если элемент уже в начале, ничего не делаем
	if l.front == i {
		return
	}

	// Отвязываем элемент от текущей позиции и связываем соседей элемента между собой
	if i.Prev != nil {
		i.Prev.Next = i.Next
	}

	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		// Если перемещаем последний элемент, обновляем указатель back
		l.back = i.Prev
	}

	// Вставляем элемент в начало списка
	i.Prev = nil     // у первого элемента нет предыдущего
	i.Next = l.front // новый первый указывает на старый первый

	if l.front != nil {
		// Обновляем ссылку у старого первого элемента
		l.front.Prev = i
	}

	// Обновляем указатель на первый элемент
	l.front = i
}
