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
		c := NewCache(3)

		c.Set("A1", 1)
		c.Set("A2", 2)
		c.Set("A3", 3)
		v, status := c.Get("A1")
		require.Equal(t, 1, v)
		require.True(t, status)
		c.Set("A4", 4)
		v, status = c.Get("A2")
		require.Equal(t, nil, v)
		require.False(t, status)

		c.Clear()

		v, status = c.Get("A1")
		require.Equal(t, nil, v)
		require.False(t, status)

		v, status = c.Get("A2")
		require.Equal(t, nil, v)
		require.False(t, status)

		v, status = c.Get("A3")
		require.Equal(t, nil, v)
		require.False(t, status)

		v, status = c.Get("A4")
		require.Equal(t, nil, v)
		require.False(t, status)
	})
}

func TestCacheMultithreading(t *testing.T) {
	_ = t
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
