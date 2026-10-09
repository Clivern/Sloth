// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/pkg/broker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingTasks captures the status calls the no-op handler makes.
type recordingTasks struct {
	db.AsyncTaskRepository
	running   []db.Id
	completed []db.Id
	err       error
}

func (f *recordingTasks) MarkRunning(id db.Id) error {
	f.running = append(f.running, id)
	return f.err
}

func (f *recordingTasks) Complete(id db.Id, result string) error {
	f.completed = append(f.completed, id)
	return nil
}

func TestUnitHandleNoop(t *testing.T) {
	t.Run("job without task id is accepted and touches nothing", func(t *testing.T) {
		tasks := &recordingTasks{}
		h := &handlers{tasks: tasks}

		err := h.HandleNoop(context.Background(), &broker.Msg{Data: []byte(`{}`)})
		require.NoError(t, err)
		assert.Empty(t, tasks.running)
		assert.Empty(t, tasks.completed)
	})

	t.Run("job with task id is marked running then completed", func(t *testing.T) {
		tasks := &recordingTasks{}
		h := &handlers{tasks: tasks}

		err := h.HandleNoop(context.Background(), &broker.Msg{Data: []byte(`{"taskId":"task-1"}`)})
		require.NoError(t, err)
		assert.Equal(t, []db.Id{"task-1"}, tasks.running)
		assert.Equal(t, []db.Id{"task-1"}, tasks.completed)
	})

	t.Run("invalid payload returns ErrInvalidPayload", func(t *testing.T) {
		h := &handlers{tasks: &recordingTasks{}}

		err := h.HandleNoop(context.Background(), &broker.Msg{Data: []byte("not json")})
		assert.ErrorIs(t, err, ErrInvalidPayload)
	})

	t.Run("mark running error is returned and job is not completed", func(t *testing.T) {
		tasks := &recordingTasks{err: errors.New("db down")}
		h := &handlers{tasks: tasks}

		err := h.HandleNoop(context.Background(), &broker.Msg{Data: []byte(`{"taskId":"task-2"}`)})
		assert.EqualError(t, err, "db down")
		assert.Empty(t, tasks.completed)
	})
}
