package framework

import (
	"context"
	"sync"

	"github.com/sidwiskers/hermes/api"
)

// ContextPool reuses handler contexts. Borrowed contexts are valid only for the
// duration of a handler call and must not be retained by applications.
type ContextPool struct {
	pool sync.Pool
}

// NewContextPool creates a handler-context pool.
func NewContextPool() *ContextPool {
	pool := &ContextPool{}
	pool.pool.New = func() any { return new(Context) }
	return pool
}

// Acquire borrows a reset context. username must already be lowercase without
// a leading @. The returned value must be paired with Release and must not
// escape the handler that owns it.
func (p *ContextPool) Acquire(ctx context.Context, bot *api.Client, update *Update, username string) *Context {
	if p == nil {
		return NewContext(ctx, bot, update, username)
	}
	item := p.pool.Get()
	value, _ := item.(*Context)
	if value == nil {
		value = new(Context)
	}
	value.reset(ctx, bot, update, username)
	return value
}

// Release clears and returns a borrowed context to the pool.
func (p *ContextPool) Release(value *Context) {
	if p == nil || value == nil {
		return
	}
	*value = Context{}
	p.pool.Put(value)
}
