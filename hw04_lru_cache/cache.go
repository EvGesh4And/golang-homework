package hw04lrucache

type Key string

type box struct {
	key   Key
	value interface{}
}

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (l *lruCache) Set(key Key, value interface{}) bool {
	if item, ok := l.items[key]; ok {
		b := item.Value.(box)
		b.value = value
		item.Value = b
		l.items[key] = item

		l.queue.MoveToFront(item)

		return true
	} else {
		if len(l.items) == l.capacity {
			item = l.queue.Back()
			b := item.Value.(box)
			keyDel := b.key
			delete(l.items, keyDel)
			l.queue.Remove(item)
		}
		l.queue.PushFront(box{key: key, value: value})
		l.items[key] = l.queue.Front()

		return false
	}
}

func (l *lruCache) Get(key Key) (interface{}, bool) {

	if item, ok := l.items[key]; ok {
		var v interface{}
		b := item.Value.(box)
		v = b.value
		l.queue.MoveToFront(item)
		return v, true
	}
	return nil, false
}

func (l *lruCache) Clear() {
	for k := range l.items {
		delete(l.items, k)
	}
	for item := l.queue.Front(); item != nil; item = item.Next {
		l.queue.Remove(item)
	}
}
