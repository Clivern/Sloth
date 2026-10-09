// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package lb

import (
	"cmp"
	"errors"
	"hash/fnv"
	"math/rand/v2"
	"slices"
	"sync"
)

var (
	ErrEmpty    = errors.New("lb: no agent instances")
	ErrStrategy = errors.New("lb: unknown strategy")
)

const (
	Random                Strategy = "random"
	RoundRobin            Strategy = "round_robin"
	LeastOutstandingTasks Strategy = "least_outstanding_tasks"
)

// Strategy is how Pick chooses among live replicas of an agent.
type Strategy string

// AgentInstance is one live replica of an agent. Map it from db.AgentInstance
// plus current outstanding-task count.
type AgentInstance struct {
	InstanceId  string
	Address     string
	Port        int
	Outstanding int
}

// Options control how Pick chooses a replica.
type Options struct {
	Strategy  Strategy
	Agent     string
	StickyKey string
}

// Picker selects a live agent instance.
type Picker struct {
	mu   sync.Mutex
	next map[string]int
	intn func(int) int
}

// New returns a picker.
func New() *Picker {
	return &Picker{
		next: make(map[string]int),
		intn: rand.IntN,
	}
}

// Pick returns one live replica of an agent.
// StickyKey pins a task to the same replica while it stays in the pool.
func (p *Picker) Pick(instances []AgentInstance, opts Options) (*AgentInstance, error) {
	if len(instances) == 0 {
		return nil, ErrEmpty
	}

	if opts.StickyKey != "" {
		picked := sticky(instances, opts.StickyKey)
		return &picked, nil
	}

	strategy := opts.Strategy
	if strategy == "" {
		strategy = Random
	}

	var picked AgentInstance
	switch strategy {
	case Random:
		picked = instances[p.intn(len(instances))]
	case RoundRobin:
		picked = p.roundRobin(opts.Agent, instances)
	case LeastOutstandingTasks:
		picked = p.leastOutstanding(instances)
	default:
		return nil, ErrStrategy
	}

	return &picked, nil
}

// Shuffle returns a copy of agent instances in random order, like Consul DNS.
func (p *Picker) Shuffle(instances []AgentInstance) []AgentInstance {
	out := slices.Clone(instances)
	for i := len(out) - 1; i > 0; i-- {
		j := p.intn(i + 1)
		out[i], out[j] = out[j], out[i]
	}

	return out
}

func (p *Picker) roundRobin(agent string, instances []AgentInstance) AgentInstance {
	ordered := ordered(instances)

	p.mu.Lock()
	i := p.next[agent]
	p.next[agent] = i + 1
	p.mu.Unlock()

	return ordered[i%len(ordered)]
}

func (p *Picker) leastOutstanding(instances []AgentInstance) AgentInstance {
	min := instances[0].Outstanding
	for _, inst := range instances[1:] {
		if inst.Outstanding < min {
			min = inst.Outstanding
		}
	}

	var tied []AgentInstance
	for _, inst := range instances {
		if inst.Outstanding == min {
			tied = append(tied, inst)
		}
	}

	return tied[p.intn(len(tied))]
}

func ordered(instances []AgentInstance) []AgentInstance {
	out := slices.Clone(instances)
	slices.SortFunc(out, func(a, b AgentInstance) int {
		return cmp.Compare(a.InstanceId, b.InstanceId)
	})

	return out
}

func sticky(instances []AgentInstance, key string) AgentInstance {
	pool := ordered(instances)
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))

	return pool[int(h.Sum32())%len(pool)]
}
