package scheduler

import (
	"context"
	"runtime"
	"sync"
	"time"

	"../job"
	"../queue"
	"github.com/jonboulle/clockwork"
)

type WaitQueue struct {
	q                                   *queue.Queue
	pl                                  Policy
	ttl, ttadj, pdelta, quota, maxQuota int64
	in, out                             chan *job.Job
	errs                                chan error
	ctx                                 context.Context
	m                                   sync.RWMutex
	clock                               clockwork.Clock
}

func (wq *WaitQueue) Start(ctx context.Context, clock clockwork.Clock, delta, ttl, ttadj, pdelta int64, multiplier float64) (<-chan *job.Job, <-chan error) {
	wq.ctx = ctx
	wq.clock = clock
	wq.ttl, wq.ttadj, wq.pdelta, wq.maxQuota = ttl, ttadj, pdelta, getJobQuota(multiplier)
	wq.quota = wq.maxQuota
	wq.in, wq.out, wq.errs = make(chan *job.Job), make(chan *job.Job), make(chan error, 1)
	ticker := time.NewTicker(time.Millisecond * time.Duration(delta))
	wq.q = queue.NewQueue(wq.clock)
	go func() {
		defer close(wq.in)
		defer close(wq.out)
		defer close(wq.errs)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				//cancel jobs
				var wg sync.WaitGroup
				for n := wq.q.Pop(); n != nil; n = wq.q.Pop() {
					j := n.(*job.Job)
					wg.Add(1)
					go func() {
						cancelJob(j)
						wg.Done()
					}()
				}
				wg.Wait()

				//drain channels
				for range wq.in {
				}
				return
			case j := <-wq.in:
				//Push in, retry and paused jobs
				if j.IsCancelled() {
					go cancelJob(j)
				} else {
					wq.q.Push(j, j.Priority)
				}
			case <-ticker.C:
				//Check Done
				expired := wq.q.RemoveIfLongerThan(ttl)
				for _, v := range expired {
					old := v.(*job.Job)
					go cancelJob(old)
				}

				//Adjust priority based on Policy
				wq.q.ChangePriorityIfLongerThan(ttadj, int(pdelta))

				//Pop jobs
				wq.PopJob()
			}
		}
	}()

	return wq.out, wq.errs
}

func (wq *WaitQueue) PopJob() {
	wq.m.Lock()
	defer wq.m.Unlock()
	l := wq.q.Len()
	if l <= 0 || wq.quota <= 0 {
		return
	}

	old := wq.q.Pop()
	if old != nil {
		wq.quota--
		go func() {
			wq.out <- old.(*job.Job)
		}()
	}
}

func (wq *WaitQueue) QueueJob(j *job.Job) {
	go func() {
		wq.in <- j
	}()
}

func (wq *WaitQueue) RunDone() {
	wq.m.Lock()
	defer wq.m.Unlock()
	wq.quota++
}

func getJobQuota(multiplier float64) int64 {
	return int64(multiplier*float64(runtime.NumCPU())) - 2
}

func cancelJob(j *job.Job) bool {
	if !j.HasProcessExited() {
		j.Stop()
		j.SetState(job.Cancelled)
		return true
	}

	return false
}
