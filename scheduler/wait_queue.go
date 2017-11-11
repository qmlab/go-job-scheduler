package scheduler

import "../queue"
import "../job"

type WaitQueue struct {
	wq        queue.Queue
	pl        Policy
	in        chan job.Job
	errs      chan error
	batchSize int
	done      <-chan struct{}
}

func (wq *WaitQueue) Start(done <-chan struct{}) (out, retry chan job.Job, errs chan error) {
	wq.in = make(chan job.Job)
	wq.done = done
	out = make(chan job.Job)
	errs = make(chan error, 1)
	retry = make(chan job.Job)
	go func() {
		defer close(wq.in)
		defer close(out)
		defer close(errs)
		defer close(retry)
		//TODO
		for {
			select {
			case <-wq.done:
				return
			case j := <-wq.in:
				//1.Check Done
				//2.Adjust priority based on Policy
				//3.Pop to out channel
				//4.Push from in, retry and paused channels
				if j.IsCancelled() {
					if j.HasProcessExited() {
						j.Stop()
					}
				} else {

				}
			}
		}
	}()

	return
}

func (wq *WaitQueue) Run(j job.Job) {
	//TODO
}
