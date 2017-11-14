package scheduler

import "context"
import "../job"

type scheduler struct {
	wq     WaitQueue
	rq     RunQueue
	es     ErrorSink
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *scheduler) Run() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	ttl, ttadj, ttr, pdelta := int64(1000*120), int64(1000*5), int64(1000*2), 1

	var errcList []<-chan error
	out, wqErr := s.wq.Start(s.ctx, ttl, ttadj, pdelta)

	errcList = append(errcList, wqErr)
	suspended, rc, rqErr := s.rq.Start(s.ctx, out, ttl, ttr)

	s.feedback(suspended, rc)

	errcList = append(errcList, rqErr)
	s.es.WaitForPipeline(s.ctx, errcList...)
}

func (s *scheduler) AddJob(j *job.Job) {
	s.wq.QueueJob(j)
}

func (s *scheduler) Close() {
	s.cancel()
}

func (s *scheduler) feedback(suspended <-chan *job.Job, rc <-chan int) {
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case j := <-suspended:
				s.wq.QueueJob(j)
			case count := <-rc:
				s.wq.SetRuncount(count)
			}
		}
	}()
}
