package scheduler

import (
	"context"
	"testing"
	"time"

	"../job"
	"github.com/stretchr/testify/assert"
)

func TestEmptyScheduler(t *testing.T) {
	s := &scheduler{}
	s.Run()
}

func TestSchedulerBasic(t *testing.T) {
	s := &scheduler{}
	s.Run()
	j := job.NewJob(context.Background(), "sleep", "1")
	s.AddJob(j)
	time.Sleep(2 * time.Second)
	s.Close()
	err := j.Wait()
	assert.Nil(t, err)
}

func TestSchedulerCancel(t *testing.T) {
	s := &scheduler{}
	s.Run()
	j := job.NewJob(context.Background(), "sleep", "3")
	s.AddJob(j)
	time.Sleep(1 * time.Second)
	s.Close()
	err := j.Wait()
	assert.NotNil(t, err)
}
