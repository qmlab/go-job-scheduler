package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"../job"
	"github.com/stretchr/testify/assert"
)

func TestEmptyScheduler(t *testing.T) {
	s := DefaultScheduler()
	s.Run()
}

func TestSchedulerBasic(t *testing.T) {
	s := DefaultScheduler()
	s.Run()
	j := job.NewJob(context.Background(), "sleep", "1")
	s.AddJob(j)
	time.Sleep(200 * time.Millisecond)
	err := j.Wait()
	assert.Nil(t, err)
}

func TestSchedulerCancel(t *testing.T) {
	s := DefaultScheduler()
	s.Run()
	j := job.NewJob(context.Background(), "sleep", "3")
	s.AddJob(j)
	time.Sleep(1 * time.Second)
	s.Close()
	err := j.Wait()
	assert.NotNil(t, err)
}

func TestSchedulerSuspend(t *testing.T) {
	s := DefaultScheduler()
	s.Run()
	j := job.NewJob(context.Background(), "sleep", "3")
	s.AddJob(j)
	time.Sleep(1 * time.Second)
	err := j.Wait()
	assert.Nil(t, err)
}

func TestSchedulerTerminate(t *testing.T) {
	s := DefaultScheduler()
	s.TTL = int64(1000)
	s.Run()
	j := job.NewJob(context.Background(), "sleep", "3")
	s.AddJob(j)
	time.Sleep(1 * time.Second)
	err := j.Wait()
	assert.NotNil(t, err)
}

func TestScheduler120(t *testing.T) {
	s := DefaultScheduler()
	s.Run()
	var wg sync.WaitGroup
	for i := 1; i < 120; i++ {
		wg.Add(1)
		j := job.NewJob(context.Background(), "echo", "2>&1")
		s.AddJob(j)
		go func() {
			err := j.Wait()
			assert.Nil(t, err)
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestScheduler1000(t *testing.T) {
	s := DefaultScheduler()
	s.Run()
	var wg sync.WaitGroup
	for i := 1; i < 1000; i++ {
		wg.Add(1)
		j := job.NewJob(context.Background(), "echo", "2>&1")
		s.AddJob(j)
		go func() {
			err := j.Wait()
			assert.Nil(t, err)
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestSchedulerStress(t *testing.T) {
	s := DefaultScheduler()
	s.TTL = int64(1000)
	s.Run()
	var wg sync.WaitGroup
	for i := 1; i < 10; i++ {
		wg.Add(1)
		j := job.NewJob(context.Background(), "cat", "/dev/random")
		s.AddJob(j)
		go func() {
			j.Wait()
			wg.Done()
		}()
	}

	wg.Wait()
}
