package scheduler

import (
	"context"

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
	ctx       context.Context
}

func (rq *RunQueue) Start(ctx context.Context, in <-chan *job.Job, ttl, ttr int64) (chan *job.Job, chan int, chan error) {
	rq.ctx = ctx
	rq.ttl, rq.ttr = ttl, ttr
	rq.suspended, rq.runcount, rq.errs = make(chan *job.Job), make(chan int), make(chan error, 1)
	go func() {
		defer close(rq.suspended)
		defer close(rq.runcount)
		defer close(rq.errs)
		for {
			select {
			case <-rq.ctx.Done():
				//drain
				for range rq.in {
				}
				for range rq.runcount {
				}
				for range rq.suspended {
				}
				return
			case j := <-rq.in:
				//1.Check Done
				expired := rq.q.RemoveIfLongerThan(ttl)
				for _, v := range expired {
					old := v.(*job.Job)
					cancelJob(old)
				}

				//2.Push & start job
				if j.State == job.Paused {
					j.Resume()
				} else {
					j.Start()
				}

				rq.q.Push(j, j.Priority)

				//3.Check & update job states
				rq.q.RemoveIf(func(v interface{}) bool {
					old := v.(*job.Job)
					if old.HasProcessExited() {
						old.State = job.Finished
						return true
					}
					if old.IsCancelled() {
						old.State = job.Cancelled
						return true
					}
					return false
				})

				//4.Pop out to wait queue according to ttr
				paused := rq.q.RemoveIfLongerThan(ttr)
				for _, v := range paused {
					old := v.(*job.Job)
					go func() {
						pauseJob(old)
						rq.suspended <- old
					}()
				}
			}
		}
	}()

	return rq.suspended, rq.runcount, rq.errs
}

func pauseJob(j *job.Job) {
	if !j.HasProcessExited() {
		j.Pause()
		j.State = job.Paused
	}
}
