// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package pool

import (
	"context"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

// Options control the size and blocking behavior of a Pool.
type Options struct {
	// Size is the maximum number of workers running at once.
	Size int

	// Nonblocking makes Submit return ErrOverload instead of waiting for a free worker.
	Nonblocking bool

	// Expiry is how long an idle worker lives before it is cleaned up.
	Expiry time.Duration
}

// Pool runs tasks on a bounded set of reused goroutines.
type Pool struct {
	client *ants.Pool
}

// New creates a Pool with the given options and returns ErrInvalidSize when Size is not positive.
func New(options Options) (*Pool, error) {
	if options.Size <= 0 {
		return nil, ErrInvalidSize
	}

	opts := []ants.Option{ants.WithNonblocking(options.Nonblocking)}
	if options.Expiry > 0 {
		opts = append(opts, ants.WithExpiryDuration(options.Expiry))
	}

	client, err := ants.NewPool(options.Size, opts...)
	if err != nil {
		return nil, err
	}

	return &Pool{client: client}, nil
}

// Submit queues task to run on a worker and returns ErrClosed or ErrOverload when it cannot.
func (p *Pool) Submit(task func()) error {
	return p.client.Submit(task)
}

// Batch runs every task on the pool and waits for the ones it started to finish.
func (p *Pool) Batch(ctx context.Context, tasks []func(context.Context)) error {
	var wg sync.WaitGroup
	var submitErr error

	for _, task := range tasks {
		if err := ctx.Err(); err != nil {
			submitErr = err
			break
		}

		wg.Add(1)
		err := p.client.Submit(func() {
			defer wg.Done()
			task(ctx)
		})
		if err != nil {
			wg.Done()
			submitErr = err
			break
		}
	}

	wg.Wait()
	return submitErr
}

// Running returns the number of workers currently running a task.
func (p *Pool) Running() int {
	return p.client.Running()
}

// Free returns the number of workers that are available to run a task.
func (p *Pool) Free() int {
	return p.client.Free()
}

// Cap returns the maximum number of workers the pool can run at once.
func (p *Pool) Cap() int {
	return p.client.Cap()
}

// Close stops the pool and waits up to timeout for running tasks to finish.
func (p *Pool) Close(timeout time.Duration) error {
	return p.client.ReleaseTimeout(timeout)
}
