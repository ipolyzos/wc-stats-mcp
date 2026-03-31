package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

// Handler executes a tool given JSON arguments and returns JSON result bytes.
type Handler func(ctx context.Context, arguments json.RawMessage) (json.RawMessage, error)

// Registry maps tool names to handlers.
type Registry struct {
	mu    sync.RWMutex
	names map[string]Handler
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{names: make(map[string]Handler)}
}

// Register adds or replaces a tool handler.
func (r *Registry) Register(name string, h Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.names[name] = h
}

// Names returns sorted registered tool names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.names))
	for n := range r.names {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Call invokes a tool by name.
func (r *Registry) Call(ctx context.Context, name string, arguments json.RawMessage) (json.RawMessage, error) {
	r.mu.RLock()
	h, ok := r.names[name]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown tool: %q", name)
	}
	return h(ctx, arguments)
}
