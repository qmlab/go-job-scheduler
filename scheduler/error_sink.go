package scheduler

import (
	"context"
	"fmt"
	"sync"
)

type ErrorSink struct {
	ctx  context.Context
	errs <-chan error
}

func (es *ErrorSink) WaitForPipeline(ctx context.Context, errs ...<-chan error) {
	es.ctx, es.errs = ctx, MergeErrors(errs...)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-es.errs:
				fmt.Printf("[Error]%s", err.Error())
			}
		}
	}()
}

// MergeErrors merges multiple channels of errors.
// Based on https://blog.golang.org/pipelines.
func MergeErrors(cs ...<-chan error) <-chan error {
	var wg sync.WaitGroup
	out := make(chan error)
	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls
	// wg.Done.
	output := func(c <-chan error) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}
	// Start a goroutine to close out once all the output goroutines
	// are done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
