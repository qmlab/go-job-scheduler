package scheduler

import (
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
	errs                    chan error
	done                    <-chan struct{}
}

func (wq *WaitQueue) Start(in chan *job.Job, ttl, ttadj int64, pdelta int, done <-chan struct{}, runcount <-chan int) (chan *job.Job, chan error) {
	wq.ttl, wq.ttadj, wq.runcount, wq.pdelta, wq.quota = ttl, ttadj, 0, pdelta, getJobQuota()
	wq.in, wq.out, wq.done, wq.errs = make(chan *job.Job), make(chan *job.Job), done, make(chan error, 1)
	go func() {
		defer close(wq.in)
		defer close(wq.out)
		defer close(wq.errs)
		for {
			select {
			case <-wq.done:
				return
			case c := <-runcount:
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
