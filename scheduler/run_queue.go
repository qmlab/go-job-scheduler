package scheduler

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"../job"
	"../queue"
)

type RunQueue struct {
	q         queue.Queue
	pl        Policy
	ttl, ttr  int64
	in        <-chan *job.Job
	suspended chan *job.Job
	errs      chan error
	runcount  chan int
	count     int32
	ctx       context.Context
}

func (rq *RunQueue) Start(ctx context.Context, in <-chan *job.Job, delta, ttl, ttr int64) (<-chan *job.Job, <-chan int, <-chan error) {
	rq.ctx = ctx
	rq.ttl, rq.ttr = ttl, ttr
	rq.in, rq.suspended, rq.runcount, rq.errs = in, make(chan *job.Job), make(chan int), make(chan error, 1)
	ticker := time.NewTicker(time.Millisecond * time.Duration(delta))
	go func() {
		defer close(rq.suspended)
		defer close(rq.runcount)
		defer close(rq.errs)
		defer ticker.Stop()
		for {
			select {
			case <-rq.ctx.Done():
				//cancel jobs
				var wg sync.WaitGroup
				for n := rq.q.Pop(); n != nil; n = rq.q.Pop() {
					j := n.(*job.Job)
					wg.Add(1)
					go func() {
						cancelJob(j)
						wg.Done()
					}()
				}
				wg.Wait()

				//drain
				for range rq.in {
				}
				for range rq.runcount {
				}
				for range rq.suspended {
				}
				return
			case j := <-rq.in:
				//Push & start job
				go func() {
					if j.GetState() == job.Paused {
						j.Resume()
						// println("resumed")
					} else {
						j.Start()
						// println("started")
					}
					rq.increaseRunCount()
					rq.q.Push(j, j.Priority)
				}()
			case <-ticker.C:
				//Check Done
				expired := rq.q.RemoveIfLongerThan(ttl)
				for _, v := range expired {
					old := v.(*job.Job)
					cancelJob(old)
				}

				//Check & update job states
				rq.q.RemoveIf(func(v interface{}) bool {
					old := v.(*job.Job)
					if old.HasProcessExited() {
						old.SetState(job.Finished)
						// println("finished")
						rq.decreaseRunCount()
						return true
					}
					if old.IsCancelled() {
						rq.decreaseRunCount()
						// println("cancelled")
						old.SetState(job.Cancelled)
						return true
					}
					return false
				})

				//Pop out to wait queue according to ttr
				paused := rq.q.RemoveIfLongerThan(ttr)
				for _, v := range paused {
					old := v.(*job.Job)
					go func() {
						pauseJob(old)
						// println("paused")
						rq.decreaseRunCount()
						rq.suspended <- old
					}()
				}
			}
		}
	}()

	return rq.suspended, rq.runcount, rq.errs
}

func (rq *RunQueue) increaseRunCount() {
	atomic.AddInt32(&rq.count, 1)
	// println("increaseRunCount.c=", rq.count)
	go func() {
		rq.runcount <- int(rq.count)
	}()
}

func (rq *RunQueue) decreaseRunCount() {
	atomic.AddInt32(&rq.count, -1)
	// println("decreaseRunCount", rq.count)
	go func() {
		rq.runcount <- int(rq.count)
	}()
}

func pauseJob(j *job.Job) {
	if !j.HasProcessExited() {
		j.Pause()
		j.SetState(job.Paused)
	}
}
