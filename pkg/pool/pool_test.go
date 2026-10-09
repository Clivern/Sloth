// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package pool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitPool(t *testing.T) {
	t.Run("rejects non positive size", func(t *testing.T) {
		_, err := New(Options{Size: 0})
		assert.ErrorIs(t, err, ErrInvalidSize)

		_, err = New(Options{Size: -1})
		assert.ErrorIs(t, err, ErrInvalidSize)
	})

	t.Run("submit runs task and capacity is reported", func(t *testing.T) {
		p, err := New(Options{Size: 3})
		require.NoError(t, err)
		defer p.Close(time.Second)

		assert.Equal(t, 3, p.Cap())

		done := make(chan struct{})
		require.NoError(t, p.Submit(func() { close(done) }))
		<-done
	})

	t.Run("nonblocking submit returns overload when full", func(t *testing.T) {
		p, err := New(Options{Size: 1, Nonblocking: true})
		require.NoError(t, err)
		defer p.Close(time.Second)

		release := make(chan struct{})
		started := make(chan struct{})
		require.NoError(t, p.Submit(func() {
			close(started)
			<-release
		}))
		<-started

		assert.ErrorIs(t, p.Submit(func() {}), ErrOverload)
		close(release)
	})

	t.Run("submit after close returns closed", func(t *testing.T) {
		p, err := New(Options{Size: 1})
		require.NoError(t, err)
		require.NoError(t, p.Close(time.Second))

		assert.ErrorIs(t, p.Submit(func() {}), ErrClosed)
	})

	t.Run("batch runs every task before returning", func(t *testing.T) {
		p, err := New(Options{Size: 4})
		require.NoError(t, err)
		defer p.Close(time.Second)

		var count atomic.Int32
		tasks := make([]func(context.Context), 20)
		for i := range tasks {
			tasks[i] = func(context.Context) {
				time.Sleep(time.Millisecond)
				count.Add(1)
			}
		}

		require.NoError(t, p.Batch(context.Background(), tasks))
		assert.Equal(t, int32(20), count.Load())
	})

	t.Run("batch passes context to tasks", func(t *testing.T) {
		p, err := New(Options{Size: 2})
		require.NoError(t, err)
		defer p.Close(time.Second)

		type key struct{}
		ctx := context.WithValue(context.Background(), key{}, "probe")

		var mu sync.Mutex
		var seen []string
		tasks := []func(context.Context){
			func(ctx context.Context) {
				mu.Lock()
				defer mu.Unlock()
				seen = append(seen, ctx.Value(key{}).(string))
			},
		}

		require.NoError(t, p.Batch(ctx, tasks))
		assert.Equal(t, []string{"probe"}, seen)
	})

	t.Run("batch stops submitting when context is cancelled", func(t *testing.T) {
		p, err := New(Options{Size: 1})
		require.NoError(t, err)
		defer p.Close(time.Second)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		var ran atomic.Int32
		tasks := []func(context.Context){
			func(context.Context) { ran.Add(1) },
		}

		assert.ErrorIs(t, p.Batch(ctx, tasks), context.Canceled)
		assert.Equal(t, int32(0), ran.Load())
	})

	t.Run("batch returns overload from nonblocking pool", func(t *testing.T) {
		p, err := New(Options{Size: 1, Nonblocking: true})
		require.NoError(t, err)
		defer p.Close(time.Second)

		release := make(chan struct{})
		started := make(chan struct{})
		require.NoError(t, p.Submit(func() {
			close(started)
			<-release
		}))
		<-started

		tasks := []func(context.Context){
			func(context.Context) {},
		}
		assert.ErrorIs(t, p.Batch(context.Background(), tasks), ErrOverload)
		close(release)
	})
}
