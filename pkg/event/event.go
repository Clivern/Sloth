// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package event

import (
	"context"
	"sync"
)

// Handler listens for an emitted event payload.
type Handler func(context.Context, any)

// Bus routes named events to registered handlers.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// Default is the application-wide event bus.
var Default = NewBus()

// NewBus creates an isolated event bus.
func NewBus() *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
	}
}

// Event is a typed named event on a bus.
type Event[E any] struct {
	bus  *Bus
	name string
}

// New declares a typed event on the default bus.
func New[E any](name string) Event[E] {
	return Event[E]{bus: Default, name: name}
}

// On registers a listener for this event.
func (e Event[E]) On(fn func(context.Context, E)) {
	e.bus.on(e.name, func(ctx context.Context, data any) {
		fn(ctx, data.(E))
	})
}

// Emit dispatches this event to all listeners.
func (e Event[E]) Emit(ctx context.Context, data E) {
	e.bus.emit(ctx, e.name, data)
}

func (b *Bus) on(name string, fn Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[name] = append(b.handlers[name], fn)
}

func (b *Bus) emit(ctx context.Context, name string, data any) {
	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers[name]...)
	b.mu.RUnlock()

	for _, fn := range handlers {
		fn(ctx, data)
	}
}
