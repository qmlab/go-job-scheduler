package scheduler

import (
	"context"
	"testing"

	"../job"
	"github.com/stretchr/testify/assert"
)

func TestEmptyScheduler(t *testing.T) {
	s := &scheduler{}
	s.Run()
}

func TestBasicScheduler(t *testing.T) {
	s := &scheduler{}
	s.Run()
	j := job.NewJob(context.Background(), "sleep", "3")
	s.AddJob(j)
	err := j.Wait()
	assert.Equal(t, nil, err)
}
