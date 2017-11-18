package scheduler

import (
	"context"
	"os"

	"../../src/github.com/jonboulle/clockwork"
	"../job"
	"../util"
)

type scheduler struct {
	wq     WaitQueue
	rq     RunQueue
	es     ErrorSink
	ctx    context.Context
	cancel context.CancelFunc

	Delta, TTL, TTP, TTS, Pstep int64
	CPUMultiplier               float64
	Clock                       clockwork.Clock
}

func DefaultScheduler() *scheduler {
	s := &scheduler{}
	s.Delta, s.TTL, s.TTP, s.TTS, s.Pstep, s.CPUMultiplier = int64(100), int64(1000*120), int64(1000*5), int64(1000*2), int64(1), 2
	s.Clock = clockwork.NewRealClock()
	return s
}

func (s *scheduler) Run() {
	pid := os.Getpid()
	util.PinToCPU(pid, 0)
	s.ctx, s.cancel = context.WithCancel(context.Background())

	var errcList []<-chan error
	out, wqErr := s.wq.Start(s.ctx, s.Clock, s.Delta, s.TTL, s.TTP, s.Pstep, s.CPUMultiplier)
	errcList = append(errcList, wqErr)

	suspended, rc, rqErr := s.rq.Start(s.ctx, s.Clock, out, s.Delta, s.TTL, s.TTS)
	errcList = append(errcList, rqErr)
	s.feedback(suspended, rc)

	s.es.WaitForPipeline(s.ctx, errcList...)
}

func (s *scheduler) AddJob(j *job.Job) {
	s.wq.QueueJob(j)
}

func (s *scheduler) Close() {
	s.cancel()
}

func (s *scheduler) feedback(suspended <-chan *job.Job, rd <-chan struct{}) {
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case j := <-suspended:
				s.wq.QueueJob(j)
			case <-rd:
				s.wq.RunDone()
			}
		}
	}()
}
