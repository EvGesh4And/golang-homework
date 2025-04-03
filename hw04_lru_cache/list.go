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

func NewList() List {
	newList := new(list)
	newList.head = &ListItem{}
	newList.tail = &ListItem{}
	return newList
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	len  int
	head *ListItem
	tail *ListItem
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.head.Next
}

func (l *list) Back() *ListItem {
	return l.tail.Prev
}

func (l *list) PushFront(v interface{}) *ListItem {
	item := &ListItem{Value: v}

	if l.head.Next == nil {
		l.head.Next = item
		l.tail.Prev = item
	} else {
		item.Next = l.head.Next
		l.head.Next.Prev = item
		l.head.Next = item
	}

	l.len++
	return item
}

func (l *list) PushBack(v interface{}) *ListItem {
	item := &ListItem{Value: v}

	if l.tail.Prev == nil {
		l.head.Next = item
		l.tail.Prev = item
	} else {
		item.Prev = l.tail.Prev
		l.tail.Prev.Next = item
		l.tail.Prev = item
	}

	l.len++
	return item
}

func (l *list) Remove(i *ListItem) {
	if i != nil {
		if i.Prev == nil {
			l.head.Next = i.Next
		} else {
			i.Prev.Next = i.Next
		}

		if i.Next == nil {
			l.tail.Prev = i.Prev
		} else {
			i.Next.Prev = i.Prev
		}

		l.len--
	}
}

func (l *list) MoveToFront(i *ListItem) {
	if i != nil && i.Prev != nil {
		l.Remove(i)
		l.PushFront(i.Value)
	}
}
