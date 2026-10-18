// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package event

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitBus(t *testing.T) {
	t.Run("On and Emit", func(t *testing.T) {
		bus := NewBus()
		ev := Event[string]{bus: bus, name: "test.event"}

		var got string
		ev.On(func(_ context.Context, payload string) {
			got = payload
		})

		ev.Emit(context.Background(), "hello")
		assert.Equal(t, "hello", got)
	})

	t.Run("Multiple handlers", func(t *testing.T) {
		bus := NewBus()
		ev := Event[int]{bus: bus, name: "count"}

		var total int
		ev.On(func(_ context.Context, n int) { total += n })
		ev.On(func(_ context.Context, n int) { total += n })

		ev.Emit(context.Background(), 2)
		assert.Equal(t, 4, total)
	})

	t.Run("Unknown event is a no-op", func(t *testing.T) {
		bus := NewBus()
		bus.emit(context.Background(), "missing", "x")
	})
}
