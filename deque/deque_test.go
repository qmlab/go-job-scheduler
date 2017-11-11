package deque

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDequeEmpty(t *testing.T) {
	l := NewDeque()
	assert.Equal(t, 0, l.Len())
}

func TestDequeBasic(t *testing.T) {
	l := NewDeque()
	assert.Equal(t, nil, l.PopFront())
	l.PushBack(2)
	l.PushBack(3)
	l.PushFront(1)
	assert.Equal(t, 3, l.Len())
	assert.Equal(t, 3, l.Back().(int))
	assert.Equal(t, 1, l.Front().(int))
	assert.Equal(t, 1, l.PopFront().(int))
	assert.Equal(t, 2, l.PopFront().(int))
	assert.Equal(t, 3, l.PopFront().(int))
	assert.Equal(t, 0, l.Len())
}

func TestDequeConcurrent(t *testing.T) {
	l := NewDeque()
	wg := &sync.WaitGroup{}
	n := 1000
	m := 1000
	f1 := func() {
		for i := 0; i < n; i++ {
			l.PushBack(i)
		}
		wg.Done()
	}
	f2 := func() {
		for i := 0; i < n; i++ {
			l.PushFront(i)
		}
		wg.Done()
	}
	g1 := func() {
		for i := 0; i < n; i++ {
			l.PopFront()
		}
		wg.Done()
	}
	g2 := func() {
		for i := 0; i < n; i++ {
			l.PopBack()
		}
		wg.Done()
	}

	for j := 0; j < n; j++ {
		wg.Add(2)
		go f1()
		go f2()
	}

	wg.Wait()
	assert.Equal(t, m*n*2, l.Len())

	for k := 0; k < n; k++ {
		wg.Add(2)
		go g1()
		go g2()
	}

	wg.Wait()
	assert.Equal(t, 0, l.Len())
}

func TestIterator(t *testing.T) {
	l := NewDeque()
	l.PushBack(0)
	l.PushBack(2)
	l.PushBack(3)
	x := 1
	it := l.GetIterator()
	for ; x > it.Current.Value.(int); it.Next() {
	}
	assert.Equal(t, 2, it.Current.Value.(int))
	l.Insert(x, it)
	assert.Equal(t, 4, l.Len())
	for i := 0; i <= 3; i++ {
		assert.Equal(t, i, l.PopFront())
	}
}
