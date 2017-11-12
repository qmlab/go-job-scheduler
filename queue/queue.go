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

type sortedSet []*sortedSetNode

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
	type pTimenode struct {
		t timenode
		p int
	}

	var toAdjust []pTimenode
	q.m.Lock()
	var newSS sortedSet
	for i := len(q.ss) - 1; i >= 0; i-- {
		ssn := q.ss[i]
		for el := ssn.nodelist.Front(); el != nil; {
			n := el.Value.(timenode)
			if time.Since(msToTime(n.ts)).Nanoseconds()/int64(time.Millisecond) > elapsedMs {
				toAdjust = append(toAdjust, pTimenode{
					t: n,
					p: ssn.p,
				})
				rm := el
				el = el.Next()
				ssn.nodelist.Remove(rm)
				q.len--
				continue
			}

			el = el.Next()
		}

		if ssn.nodelist.Len() > 0 {
			newSS = append(newSS, ssn)
		}
	}

	sort.Sort(newSS)
	q.ss = newSS

	q.m.Unlock()
	for _, n := range toAdjust {
		q.Insert(n.t, n.p+delta)
	}
}

func msToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

// Insert puts the job into correct location based on timestamp
func (q *Queue) Insert(node timenode, priority int) error {
	q.m.Lock()
	defer q.m.Unlock()

	i := q.ss.Find(priority)
	var ssn *sortedSetNode
	if i >= 0 {
		ssn = q.ss[i]
		el := ssn.nodelist.Front()
		for ; el != nil; el = el.Next() {
			if el.Value.(timenode).ts > node.ts {
				ssn.nodelist.InsertBefore(node, el)
			}
		}
		if el == nil {
			ssn.nodelist.PushBack(node)
		}
	} else {
		ssn = &sortedSetNode{
			p:        priority,
			nodelist: list.New(),
		}
		q.ss = append(q.ss, ssn)
		sort.Sort(q.ss)
		ssn.nodelist.PushBack(node)
	}

	q.len++
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
	var ssn *sortedSetNode
	if i >= 0 {
		ssn = q.ss[i]
	} else {
		ssn = &sortedSetNode{
			p:        priority,
			nodelist: list.New(),
		}
		q.ss = append(q.ss, ssn)
		sort.Sort(q.ss)
	}

	ssn.nodelist.PushBack(node)
	q.len++

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
}

func (q *Queue) Len() int {
	q.m.RLock()
	defer q.m.RUnlock()
	return q.len
}
