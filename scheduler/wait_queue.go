package scheduler

import (
	"context"
	"runtime"

	"../job"
	"../queue"
)

type WaitQueue struct {
	q                       queue.Queue
	pl                      Policy
	ttl, ttadj              int64
	runcount, pdelta, quota int
	in, out                 chan *job.Job
	rc                      chan int
	errs                    chan error
	ctx                     context.Context
}

func (wq *WaitQueue) Start(ctx context.Context, ttl, ttadj int64, pdelta int) (chan *job.Job, chan error) {
	wq.ctx = ctx
	wq.ttl, wq.ttadj, wq.runcount, wq.pdelta, wq.quota = ttl, ttadj, 0, pdelta, getJobQuota()
	wq.in, wq.out, wq.errs, wq.rc = make(chan *job.Job), make(chan *job.Job), make(chan error, 1), make(chan int)
	go func() {
		defer close(wq.in)
		defer close(wq.out)
		defer close(wq.errs)
		defer close(wq.rc)
		for {
			select {
			case <-ctx.Done():
				//drain channels
				for range wq.rc {
				}
				for range wq.in {
				}
				return
			case c := <-wq.rc:
				wq.runcount = c
			case j := <-wq.in:
				//1.Check Done
				expired := wq.q.RemoveIfLongerThan(ttl)
				for _, v := range expired {
					old := v.(*job.Job)
					cancelJob(old)
				}

				//2.Adjust priority based on Policy
				wq.q.ChangePriorityIfLongerThan(ttadj, pdelta)

				//3.Pop to out channel
				for i := wq.quota - wq.runcount; i > 0 && wq.q.Len() > 0; i-- {
					old := wq.q.Pop().(*job.Job)
					old.State = job.Starting
					wq.out <- old
				}

				//4.Push in, retry and paused jobs
				if j.IsCancelled() {
					cancelJob(j)
				} else {
					wq.q.Push(j, j.Priority)
				}
			}
		}
	}()

	return wq.out, wq.errs
}

func (wq *WaitQueue) QueueJob(j *job.Job) {
	wq.in <- j
}

func (wq *WaitQueue) SetRuncount(count int) {
	wq.rc <- count
}

func getJobQuota() int {
	return runtime.NumCPU() - 1
}

func cancelJob(j *job.Job) {
	go func() {
		if !j.HasProcessExited() {
			j.Stop()
			j.State = job.Cancelled
		}
	}()
}
