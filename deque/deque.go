package deque

import (
	"fmt"
	"sync"
)

type Node struct {
	Value interface{}
	next  *Node
	prev  *Node
}

type Deque struct {
	Cap int

	len   int
	m     *sync.RWMutex
	front *Node
	back  *Node
}

func NewDeque() *Deque {
	l := &Deque{
		m: &sync.RWMutex{},
	}

	return l
}

type Iterator struct {
	Current *Node
}

func (it *Iterator) End() bool {
	return it.Current == nil
}

func (it *Iterator) Next() {
	if it.End() {
		return
	}

	it.Current = it.Current.next
}

func (l *Deque) GetIterator() Iterator {
	return Iterator{
		Current: l.front,
	}
}

func (l *Deque) Insert(v interface{}, it Iterator) {
	l.m.Lock()
	defer l.m.Unlock()
	if it.Current == nil {
		node := &Node{
			Value: v,
			prev:  l.back,
		}
		if l.back != nil {
			l.back.next = node
		} else {
			l.front = node
		}

		l.back = node
	} else {
		node := &Node{
			Value: v,
			prev:  it.Current.prev,
			next:  it.Current,
		}

		if it.Current.prev != nil {
			it.Current.prev.next = node
		}

		it.Current.prev = node
		if it.Current == l.front {
			l.front = node
		}
	}

	l.len++
}

func (l *Deque) PushFront(v interface{}) error {
	l.m.Lock()
	defer l.m.Unlock()
	if l.Cap > 0 && l.len >= l.Cap {
		return fmt.Errorf("Deque is full. Cap is %d", l.len)
	}
	n := &Node{
		Value: v,
	}
	if l.front == nil {
		l.front = n
		l.back = n
	} else {
		n.next = l.front
		l.front.prev = n
		l.front = n
	}

	l.len++
	return nil
}

func (l *Deque) PushBack(v interface{}) error {
	l.m.Lock()
	defer l.m.Unlock()
	if l.Cap > 0 && l.len == l.Cap {
		return fmt.Errorf("Deque is full. Cap is %d", l.len)
	}
	n := &Node{
		Value: v,
	}
	if l.back == nil {
		l.front = n
		l.back = n
	} else {
		n.prev = l.back
		l.back.next = n
		l.back = n
	}

	l.len++
	return nil
}

func (l *Deque) PopFront() interface{} {
	l.m.Lock()
	defer l.m.Unlock()
	if l.len == 0 {
		return nil
	}
	n := l.front
	l.front = n.next
	if l.front == nil {
		l.back = nil
	} else {
		l.front.prev = nil
	}

	l.len--
	return n.Value
}

func (l *Deque) PopBack() interface{} {
	l.m.Lock()
	defer l.m.Unlock()
	if l.len == 0 {
		return nil
	}
	n := l.back
	l.back = n.prev
	if l.back == nil {
		l.front = nil
	} else {
		l.back.next = nil
	}

	l.len--
	return n.Value
}

func (l *Deque) Front() interface{} {
	l.m.RLock()
	defer l.m.RUnlock()
	if l.front == nil {
		return nil
	}
	return l.front.Value
}

func (l *Deque) Back() interface{} {
	l.m.RLock()
	defer l.m.RUnlock()
	if l.back == nil {
		return nil
	}
	return l.back.Value
}

func (l *Deque) Len() int {
	l.m.RLock()
	defer l.m.RUnlock()
	return l.len
}
