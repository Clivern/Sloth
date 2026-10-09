// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package lb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitPicker(t *testing.T) {
	a := AgentInstance{InstanceId: "invoice-extractor-1", Address: "10.0.12.41", Port: 8080, Outstanding: 2}
	b := AgentInstance{InstanceId: "invoice-extractor-2", Address: "10.0.12.42", Port: 8080, Outstanding: 0}
	c := AgentInstance{InstanceId: "invoice-extractor-3", Address: "10.0.12.43", Port: 8080, Outstanding: 0}

	t.Run("empty pool", func(t *testing.T) {
		_, err := New().Pick(nil, Options{})
		assert.ErrorIs(t, err, ErrEmpty)
	})

	t.Run("single instance", func(t *testing.T) {
		got, err := New().Pick([]AgentInstance{a}, Options{})
		require.NoError(t, err)
		assert.Equal(t, "invoice-extractor-1", got.InstanceId)
		assert.Equal(t, "10.0.12.41", got.Address)
		assert.Equal(t, 8080, got.Port)
	})

	t.Run("unknown strategy", func(t *testing.T) {
		_, err := New().Pick([]AgentInstance{a}, Options{Strategy: "weighted"})
		assert.ErrorIs(t, err, ErrStrategy)
	})

	t.Run("random uses injected index", func(t *testing.T) {
		p := New()
		p.intn = func(n int) int {
			assert.Equal(t, 3, n)

			return 1
		}
		got, err := p.Pick([]AgentInstance{a, b, c}, Options{Strategy: Random})
		require.NoError(t, err)
		assert.Equal(t, "invoice-extractor-2", got.InstanceId)
	})

	t.Run("round robin cycles in instance id order", func(t *testing.T) {
		p := New()
		pool := []AgentInstance{c, a, b}
		var ids []string
		for range 4 {
			got, err := p.Pick(pool, Options{Strategy: RoundRobin, Agent: "invoice-extractor"})
			require.NoError(t, err)
			ids = append(ids, got.InstanceId)
		}

		assert.Equal(t, []string{
			"invoice-extractor-1",
			"invoice-extractor-2",
			"invoice-extractor-3",
			"invoice-extractor-1",
		}, ids)
	})

	t.Run("least outstanding picks the lightest", func(t *testing.T) {
		p := New()
		p.intn = func(int) int { return 0 }
		got, err := p.Pick([]AgentInstance{a, b}, Options{Strategy: LeastOutstandingTasks})
		require.NoError(t, err)
		assert.Equal(t, "invoice-extractor-2", got.InstanceId)
	})

	t.Run("least outstanding breaks ties at random", func(t *testing.T) {
		p := New()
		p.intn = func(n int) int {
			assert.Equal(t, 2, n)

			return 1
		}
		got, err := p.Pick([]AgentInstance{a, b, c}, Options{Strategy: LeastOutstandingTasks})
		require.NoError(t, err)
		assert.Equal(t, "invoice-extractor-3", got.InstanceId)
	})

	t.Run("sticky key is stable", func(t *testing.T) {
		p := New()
		pool := []AgentInstance{a, b, c}
		first, err := p.Pick(pool, Options{StickyKey: "task-1"})
		require.NoError(t, err)
		second, err := p.Pick(pool, Options{Strategy: Random, StickyKey: "task-1"})
		require.NoError(t, err)
		assert.Equal(t, first.InstanceId, second.InstanceId)
	})

	t.Run("sticky follows remaining members", func(t *testing.T) {
		p := New()
		full, err := p.Pick([]AgentInstance{a, b, c}, Options{StickyKey: "task-9"})
		require.NoError(t, err)

		remaining := make([]AgentInstance, 0, 2)
		for _, inst := range []AgentInstance{a, b, c} {
			if inst.InstanceId != full.InstanceId {
				remaining = append(remaining, inst)
			}
		}

		got, err := p.Pick(remaining, Options{StickyKey: "task-9"})
		require.NoError(t, err)
		assert.NotEqual(t, full.InstanceId, got.InstanceId)
		assert.Contains(t, []string{remaining[0].InstanceId, remaining[1].InstanceId}, got.InstanceId)
	})
}

func TestUnitShuffle(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		assert.Empty(t, New().Shuffle(nil))
	})

	t.Run("does not mutate input", func(t *testing.T) {
		in := []AgentInstance{
			{InstanceId: "invoice-extractor-1"},
			{InstanceId: "invoice-extractor-2"},
			{InstanceId: "invoice-extractor-3"},
		}
		p := New()
		p.intn = func(n int) int { return n - 1 }
		out := p.Shuffle(in)
		assert.Equal(t, "invoice-extractor-1", in[0].InstanceId)
		assert.Len(t, out, 3)
	})
}
