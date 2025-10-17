package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic", func(t *testing.T) {
		// Тест на выталкивание элементов из-за размера очереди
		c := NewCache(3)

		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)

		// Все элементы должны быть в кэше
		val, ok := c.Get("a")
		require.True(t, ok)
		require.Equal(t, 1, val)

		val, ok = c.Get("b")
		require.True(t, ok)
		require.Equal(t, 2, val)

		val, ok = c.Get("c")
		require.True(t, ok)
		require.Equal(t, 3, val)

		// Добавляем 4-й элемент - первый должен вытолкнуться
		c.Set("d", 4)

		// Элемент "a" должен быть вытолкнут
		_, ok = c.Get("a")
		require.False(t, ok)

		// Остальные элементы должны быть в кэше
		val, ok = c.Get("b")
		require.True(t, ok)
		require.Equal(t, 2, val)

		val, ok = c.Get("c")
		require.True(t, ok)
		require.Equal(t, 3, val)

		val, ok = c.Get("d")
		require.True(t, ok)
		require.Equal(t, 4, val)
	})

	t.Run("purge least recently used", func(t *testing.T) {
		c := NewCache(3)

		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)

		// Обращаемся к элементам в разном порядке
		c.Get("a")     // a становится самым свежим
		c.Set("b", 20) // b обновляется и становится самым свежим
		c.Get("a")     // a снова становится самым свежим

		// Добавляем новый элемент - c должен вытолкнуться
		c.Set("d", 4)
		_, ok := c.Get("c")
		require.False(t, ok)

		// Остальные элементы должны быть в кэше
		val, ok := c.Get("a")
		require.True(t, ok)
		require.Equal(t, 1, val)

		val, ok = c.Get("b")
		require.True(t, ok)
		require.Equal(t, 20, val)

		val, ok = c.Get("d")
		require.True(t, ok)
		require.Equal(t, 4, val)
	})

	t.Run("clear cache", func(t *testing.T) {
		c := NewCache(3)

		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)

		c.Clear()

		// Все элементы должны быть удалены
		_, ok := c.Get("a")
		require.False(t, ok)

		_, ok = c.Get("b")
		require.False(t, ok)

		_, ok = c.Get("c")
		require.False(t, ok)
	})
}

func TestCacheMultithreading(t *testing.T) {
	t.Skip() // Remove me if task with asterisk completed.

	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}
