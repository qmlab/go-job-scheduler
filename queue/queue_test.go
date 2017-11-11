package queue

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"../job"

	"github.com/stretchr/testify/assert"
)

func TestEmptyQueue(t *testing.T) {
	q := NewQueue()
	assert.Equal(t, 0, q.Len())
}

func TestSimpleQueue(t *testing.T) {
	q := NewQueue()
	n := 10
	for i := 0; i < n; i++ {
		j := job.NewJob(context.Background(), "sleep", "1")
		j.Id = i
		q.Push(j, 1)
		assert.Equal(t, i+1, q.Len())
	}
	for i := 0; i < n; i++ {
		j := job.NewJob(context.Background(), "sleep", "1")
		j.Id = i * 2
		q.Push(j, 2)
	}
	assert.Equal(t, n*2, q.Len())

	for i := 0; i < n; i++ {
		v := q.Pop().(*job.Job)
		assert.Equal(t, i*2, v.Id)
	}

	assert.Equal(t, n, q.Len())
	for i := 0; i < n; i++ {
		v := q.Pop().(*job.Job)
		assert.Equal(t, i, v.Id)
	}
}

func TestConcurrentQueue(t *testing.T) {
	q := NewQueue()
	n := 10
	wg := sync.WaitGroup{}
	wg.Add(n * 2)
	go func() {
		for i := 0; i < n; i++ {
			j := job.NewJob(context.Background(), "sleep", "1")
			j.Id = i
			q.Push(j, 1)
			wg.Done()
		}
	}()

	go func() {
		for i := 0; i < n; i++ {
			j := job.NewJob(context.Background(), "sleep", "1")
			j.Id = i * 2
			q.Push(j, 2)
			wg.Done()
		}
	}()

	wg.Wait()
	assert.Equal(t, n*2, q.Len())
	wg.Add(n * 2)
	go func() {
		for i := 0; i < n; i++ {
			q.Pop()
			wg.Done()
		}
	}()

	go func() {
		for i := 0; i < n; i++ {
			q.Pop()
			wg.Done()
		}
	}()

	wg.Wait()
	assert.Equal(t, 0, q.Len())
}

func TestPriorityChange(t *testing.T) {
	q := NewQueue()
	n := 10
	for i := 0; i < n; i++ {
		j := job.NewJob(context.Background(), "sleep", "1")
		j.Id = i
		j.Priority = 1
		q.Push(j, 1)
		assert.Equal(t, i+1, q.Len())
	}

	time.Sleep(1 * time.Second)
	for i := 0; i < n; i++ {
		j := job.NewJob(context.Background(), "sleep", "1")
		j.Id = i * 10
		j.Priority = 2
		q.Push(j, 2)
	}
	assert.Equal(t, n*2, q.Len())

	q.ChangePriorityIfLongerThan(500, 2)
	fmt.Println("%v", q.ss[1].nodelist.Front().Next().Value.(timenode).value.(*job.Job).Id)

	for i := 0; i < n; i++ {
		v := q.Pop().(*job.Job)
		assert.Equal(t, i*10, v.Id)
	}

	assert.Equal(t, n, q.Len())
	for i := 0; i < n; i++ {
		v := q.Pop().(*job.Job)
		assert.Equal(t, i, v.Id)
	}
}
