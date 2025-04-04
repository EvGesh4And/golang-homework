package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})
}

func TestListScript(t *testing.T) {
	t.Run("front", func(t *testing.T) {
		l := NewList()

		l.PushFront(11)
		node3 := l.PushFront(12)
		node4 := l.PushFront(13)
		l.PushFront(14)
		l.MoveToFront(node4)
		l.Remove(node3)

		expected := []int{13, 14, 11}
		actual := []int{}
		for i := l.Front(); i != nil; i = i.Next {
			actual = append(actual, i.Value.(int))
		}
		require.Equal(t, expected, actual)
	})

	t.Run("back", func(t *testing.T) {
		l := NewList()

		l.PushBack(11)
		node3 := l.PushBack(12)
		node4 := l.PushBack(13)
		l.PushBack(14)
		l.MoveToFront(node4)
		l.Remove(node3)

		expected := []int{13, 11, 14}
		actual := []int{}
		for i := l.Front(); i != nil; i = i.Next {
			actual = append(actual, i.Value.(int))
		}
		require.Equal(t, expected, actual)
	})

	t.Run("front&back", func(t *testing.T) {
		expected := []int{3, 4, 5, 2, 10, 10, 3, 100}
		actual := make([]int, 0, len(expected))
		l := NewList()
		node := l.PushBack(2)
		node2 := l.PushBack(3)
		actual = append(actual, l.Back().Value.(int))
		l.PushBack(4)
		l.PushFront(5)
		actual = append(actual, l.Back().Value.(int))
		actual = append(actual, l.Front().Value.(int))
		l.MoveToFront(node)
		actual = append(actual, l.Front().Value.(int))
		node3 := l.PushFront(10)
		actual = append(actual, l.Front().Value.(int))
		l.Remove(node2)
		actual = append(actual, l.Front().Value.(int))
		actual = append(actual, node2.Value.(int))
		l.PushBack(100)
		actual = append(actual, l.Back().Value.(int))
		require.Equal(t, expected, actual)
		l.Remove(node3)

		expected = []int{2, 5, 4, 100}
		actual = []int{}
		for i := l.Front(); i != nil; i = i.Next {
			actual = append(actual, i.Value.(int))
		}
		require.Equal(t, expected, actual)

		expected = []int{100, 4, 5, 2}
		actual = []int{}
		for i := l.Back(); i != nil; i = i.Prev {
			actual = append(actual, i.Value.(int))
		}
		require.Equal(t, expected, actual)
	})
}
