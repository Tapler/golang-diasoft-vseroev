package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int               // максимальное количество элементов в кэше
	queue    List              // очередь элементов (самые свежие в начале)
	items    map[Key]*ListItem // словарь для быстрого поиска: ключ -> элемент очереди
}

type cacheItem struct {
	key   Key
	value interface{}
}

// NewCache создаёт новый LRU-кэш с заданной ёмкостью.
func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

// Set добавляет значение в кэш по ключу.
// Возвращаемое значение - флаг, присутствовал ли элемент в кэше.
func (c *lruCache) Set(key Key, value interface{}) bool {
	if item, ok := c.items[key]; ok {
		// Случай 1: Элемент уже есть в кэше
		// Обновляем его значение и перемещаем в начало очереди
		item.Value = cacheItem{key: key, value: value}
		c.queue.MoveToFront(item)
		return true
	}

	// Случай 2: Элемента нет в кэше - добавляем новый
	// Добавляем в начало очереди
	item := c.queue.PushFront(cacheItem{key: key, value: value})
	// Регистрируем в словаре для быстрого поиска
	c.items[key] = item

	// Проверяем, не превышена ли ёмкость кэша
	if c.queue.Len() > c.capacity {
		// Удаляем самый старый элемент (последний в очереди)
		lastItem := c.queue.Back()
		if lastItem != nil {
			// Удаляем из очереди
			c.queue.Remove(lastItem)
			// Удаляем из словаря по ключу
			lastCacheItem := lastItem.Value.(cacheItem)
			delete(c.items, lastCacheItem.key)
		}
	}

	return false
}

// Get получает значение из кэша по ключу.
// Возвращает значение и true, если элемент найден, иначе nil и false.
// При успешном получении элемент перемещается в начало очереди.
func (c *lruCache) Get(key Key) (interface{}, bool) {
	if item, ok := c.items[key]; ok {
		// Элемент найден в кэше
		// Перемещаем в начало очереди
		c.queue.MoveToFront(item)
		// Извлекаем и возвращаем значение
		cacheItem := item.Value.(cacheItem)
		return cacheItem.value, true
	}

	// Элемент не найден в кэше
	return nil, false
}

// Clear полностью очищает кэш, удаляя все элементы.
func (c *lruCache) Clear() {
	// обход всех элементов
	for key := range c.items {
		delete(c.items, key)
	}

	// удаление всех элементов списка
	for c.queue.Len() > 0 {
		c.queue.Remove(c.queue.Front())
	}
}
