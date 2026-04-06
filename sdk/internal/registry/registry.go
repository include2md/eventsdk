package registry

import (
	"context"
	"sync"
)

type Handler func(ctx context.Context, payload []byte) error

type Registry interface {
	Register(eventType string, handler Handler)
	Get(eventType string) (Handler, bool)
}

type SafeRegistry struct {
	mu       sync.RWMutex
	handlers map[string]Handler
}

func New() *SafeRegistry {
	return &SafeRegistry{handlers: make(map[string]Handler)}
}

func (r *SafeRegistry) Register(eventType string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[eventType] = handler
}

func (r *SafeRegistry) Get(eventType string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.handlers[eventType]
	return h, ok
}
