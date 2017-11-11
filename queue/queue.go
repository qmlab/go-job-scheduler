package queue

import (
	"container/list"
	"sort"
	"sync"
	"time"
)

// Concurrent sorted queue

type Queue struct {
	m   sync.RWMutex
	len int
	ss  sortedSet
}

type sortedSet []sortedSetNode

func (s sortedSet) Len() int           { return len(s) }
func (s sortedSet) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sortedSet) Less(i, j int) bool { return s[i].p < s[j].p }
func (s sortedSet) Find(p int) int {
	for i, nq := range s {
		if p == nq.p {
			return i
		}
	}

	return -1
}

type sortedSetNode struct {
	p        int
	nodelist *list.List
}

type timenode struct {
	ts    int64
	value interface{}
}

func NewQueue() *Queue {
	var q Queue
	q.m = sync.RWMutex{}
	return &q
}

func (q *Queue) ChangePriorityIfLongerThan(elapsedMs int64, delta int) {
	//TODO
	type pTimenode struct {
		t timenode
		p int
	}
	var toAdjust []pTimenode
	q.m.RLock()
	for i := len(q.ss) - 1; i >= 0; i-- {
		ssn := q.ss[i]
		for el := ssn.nodelist.Front(); el != nil; el = el.Next() {
			n := el.Value.(timenode)
			if time.Since(msToTime(n.ts)).Nanoseconds()/int64(time.Millisecond) > elapsedMs {
				toAdjust = append(toAdjust, pTimenode{
					t: n,
					p: ssn.p,
				})
			}
		}
	}

	q.m.RUnlock()
	q.m.Lock()
	for _, n := range toAdjust {
		q.Insert(n.t, n.p+delta)
	}

	// var sortedSetNode []timenode
	// for p, dq := range q.lm {
	// 	for node := dq.Front().(timenode); dq.Len() > 0 && time.Since(msToTime(node.ts)).Nanoseconds()/int64(time.Millisecond) > elapsedMs; {
	// 		j := node.value.(*job.Job)
	// 		j.Priority = p + delta
	// 		dq.PopFront()
	// 		sortedSetNode = append(sortedSetNode, node)
	// 		q.m.Lock()
	// 		if dq.Len() == 0 {
	// 			delete(q.lm, p)
	// 			i := heap.Pop(q.h).(int)
	// 			println("heap.Pop(q.h):", i)
	// 		}
	// 		q.m.Unlock()
	// 	}
	// }
	//
	// for _, node := range sortedSetNode {
	// 	j := node.value.(*job.Job)
	// 	q.Insert(node, j.Priority)
	// }
}

func msToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

// Insert puts the job into correct location based on timestamp
func (q *Queue) Insert(node timenode, priority int) error {
	q.m.Lock()
	defer q.m.Unlock()
	//TODO

	i := q.ss.Find(priority)
	var ssn sortedSetNode
	if i >= 0 {
		ssn = q.ss[i]
		for el := ssn.nodelist.Front(); el != nil; el = el.Next() {
			if el.Value.(timenode).ts > node.ts {
				ssn.nodelist.InsertBefore(node, el)
			}
		}
	} else {
		ssn = sortedSetNode{}
		ssn.nodelist = list.New()
		q.ss = append(q.ss, ssn)
		sort.Sort(q.ss)
		ssn.nodelist.PushBack(node)
	}

	// if _, ok := q.lm[priority]; !ok {
	// 	dq := deque.NewDeque()
	// 	q.lm[priority] = dq
	// 	heap.Push(q.h, priority)
	// }
	//
	// l := q.lm[priority]
	// it := l.GetIterator()
	// for ; !it.End() && node.ts > it.Current.Value.(timenode).ts; it.Next() {
	// }
	//
	// l.Insert(node, it)
	// q.len++
	return nil
}

func (q *Queue) Push(j interface{}, priority int) error {
	if j == nil {
		return nil
	}

	node := timenode{
		ts:    time.Now().UTC().UnixNano() / int64(time.Millisecond),
		value: j,
	}

	q.m.Lock()
	defer q.m.Unlock()
	i := q.ss.Find(priority)
	var ssn sortedSetNode
	if i >= 0 {
		ssn = q.ss[i]
	} else {
		ssn = sortedSetNode{}
		ssn.nodelist = list.New()
		q.ss = append(q.ss, ssn)
		sort.Sort(q.ss)
	}

	ssn.nodelist.PushBack(node)
	q.len++

	// if _, ok := q.lm[priority]; !ok {
	// 	dq := deque.NewDeque()
	// 	q.lm[priority] = dq
	// 	heap.Push(q.h, priority)
	// }
	//
	// l := q.lm[priority]
	// l.PushBack(node)
	// q.len++
	return nil
}

func (q *Queue) Pop() interface{} {
	q.m.Lock()
	defer q.m.Unlock()

	if q.len == 0 {
		return nil
	}

	l := len(q.ss)
	nq := q.ss[l-1]
	el := nq.nodelist.Front()
	n := el.Value.(timenode)
	nq.nodelist.Remove(el)
	if nq.nodelist.Len() == 0 {
		q.ss = q.ss[:l-1]
	}

	q.len--
	return n.value

	// p := q.h.Peek().(int)
	// dq := q.lm[p]
	// node := dq.PopFront()
	// if dq.Len() == 0 {
	// 	delete(q.lm, p)
	//
	// 	// Bug
	// 	heap.Pop(q.h)
	// }
	//
	// q.len--
	// return node.(timenode).value
}

func (q *Queue) Peek() interface{} {
	if q.len == 0 {
		return nil
	}

	q.m.RLock()
	defer q.m.RUnlock()

	l := len(q.ss)
	nq := q.ss[l-1]
	el := nq.nodelist.Front()
	n := el.Value.(timenode)

	return n.value

	// p := q.h.Peek().(int)
	// dq := q.lm[p]
	// return dq.Front().(timenode).value
}

func (q *Queue) Len() int {
	q.m.RLock()
	defer q.m.RUnlock()
	return q.len
}

//
// type intHeap []int
//
// func (h intHeap) Len() int           { return len(h) }
// func (h intHeap) Less(i, j int) bool { return h[i] > h[j] }
// func (h intHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
//
// func (h *intHeap) Push(x interface{}) {
// 	// Push and Pop use pointer receivers because they modify the slice's length,
// 	// not just its contents.
// 	*h = append(*h, x.(int))
// }
//
// func (h *intHeap) Pop() interface{} {
// 	old := *h
// 	n := len(old)
// 	x := old[n-1]
// 	*h = old[0 : n-1]
// 	return x
// }
//
// func (h *intHeap) Peek() interface{} {
// 	v := heap.Pop(h)
// 	heap.Push(h, v)
// 	return v
// }

// type timeHeap []timenode
//
// func (h timeHeap) Len() int           { return len(h) }
// func (h timeHeap) Less(i, j int) bool { return h[i].ts < h[j].ts }
// func (h timeHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
//
// func (h *timeHeap) Push(x interface{}) {
// 	// Push and Pop use pointer receivers because they modify the slice's length,
// 	// not just its contents.
// 	*h = append(*h, x.(timenode))
// }
//
// func (h *timeHeap) Pop() interface{} {
// 	old := *h
// 	n := len(old)
// 	x := old[0]
// 	*h = old[1:n]
// 	return x
// }
//
// func (h *timeHeap) Peek() interface{} {
// 	old := *h
// 	x := old[0]
// 	return x
// }
