package scheduler

import "../queue"
import "../job"

type RunQueue struct {
	q         queue.Queue
	pl        Policy
	ttl, ttr  int64
	in        <-chan *job.Job
	suspended chan *job.Job
	errs      chan error
	done      <-chan struct{}
	runcount  chan int
}

func (wq *RunQueue) Start(in <-chan *job.Job, ttl, ttr int64, done <-chan struct{}) (chan *job.Job, chan int, chan error) {
	wq.ttl, wq.ttr = ttl, ttr
	wq.suspended, wq.runcount, wq.done, wq.errs = make(chan *job.Job), make(chan int), done, make(chan error, 1)
	go func() {
		defer close(wq.suspended)
		defer close(wq.runcount)
		defer close(wq.errs)
		for {
			select {
			case <-wq.done:
				return
			case j := <-wq.in:
				//1.Check Done
				expired := wq.q.RemoveIfLongerThan(ttl)
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

				wq.q.Push(j, j.Priority)

				//3.Check & update job states
				wq.q.RemoveIf(func(v interface{}) bool {
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
				paused := wq.q.RemoveIfLongerThan(ttr)
				for _, v := range paused {
					old := v.(*job.Job)
					go func() {
						pauseJob(old)
						wq.suspended <- old
					}()
				}
			}
		}
	}()

	return wq.suspended, wq.runcount, wq.errs
}

func pauseJob(j *job.Job) {
	if !j.HasProcessExited() {
		j.Pause()
		j.State = job.Paused
	}
}
