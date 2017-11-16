package scheduler

import (
	"context"
	"sync"
	"time"

	"../job"
	"../queue"
	"github.com/jonboulle/clockwork"
)

type RunQueue struct {
	q         *queue.Queue
	pl        Policy
	ttl, ttr  int64
	in        <-chan *job.Job
	suspended chan *job.Job
	errs      chan error
	rundone   chan struct{}
	count     int32
	ctx       context.Context
	clock     clockwork.Clock
}

func (rq *RunQueue) Start(ctx context.Context, clock clockwork.Clock, in <-chan *job.Job, delta, ttl, ttr int64) (<-chan *job.Job, <-chan struct{}, <-chan error) {
	rq.ctx = ctx
	rq.clock = clock
	rq.ttl, rq.ttr = ttl, ttr
	rq.in, rq.suspended, rq.rundone, rq.errs = in, make(chan *job.Job), make(chan struct{}), make(chan error, 1)
	ticker := time.NewTicker(time.Millisecond * time.Duration(delta))
	rq.q = queue.NewQueue(rq.clock)
	go func() {
		defer close(rq.suspended)
		defer close(rq.rundone)
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
				for range rq.rundone {
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
					rq.q.Push(j, j.Priority)
				}()
			case <-ticker.C:
				//Check Done
				expired := rq.q.RemoveIfLongerThan(ttl)
				for _, v := range expired {
					old := v.(*job.Job)
					go cancelJob(old)
					rq.runDone()
				}

				//Check & update job states
				rq.q.RemoveIf(func(v interface{}) bool {
					old := v.(*job.Job)
					if old.HasProcessExited() {
						old.SetState(job.Finished)
						// println("finished")
						rq.runDone()
						return true
					}
					if old.IsCancelled() {
						// println("cancelled")
						rq.runDone()
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
						// println("paused")
						pauseJob(old)
						rq.runDone()
						rq.suspended <- old
					}()
				}
			}
		}
	}()

	return rq.suspended, rq.rundone, rq.errs
}

func (rq *RunQueue) runDone() {
	go func() {
		rq.rundone <- struct{}{}
	}()
}

func pauseJob(j *job.Job) {
	if !j.HasProcessExited() {
		j.Pause()
		j.SetState(job.Paused)
	}
}
