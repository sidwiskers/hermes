package runtime

import (
	"context"
	"sync"

	telegram "github.com/sidwiskers/hermes/types"
)

// UpdateHandler processes one update. Dispatcher owns concurrency, not policy.
type UpdateHandler func(context.Context, *telegram.Update)

// Dispatcher provides bounded concurrent update execution and graceful draining.
type Dispatcher struct {
	slots   chan struct{}
	mu      sync.Mutex
	drained *sync.Cond
	active  int
	handler UpdateHandler
}

func NewDispatcher(limit int, handler UpdateHandler) *Dispatcher {
	if limit < 1 {
		limit = 1
	}
	dispatcher := &Dispatcher{slots: make(chan struct{}, limit), handler: handler}
	dispatcher.drained = sync.NewCond(&dispatcher.mu)
	return dispatcher
}

// Queue schedules an update. wait controls whether backpressure blocks or fails fast.
func (d *Dispatcher) Queue(ctx context.Context, update *telegram.Update, wait bool) bool {
	if update == nil {
		return true
	}
	if d == nil || d.handler == nil {
		return false
	}
	release, ok := d.Reserve(ctx, wait)
	if !ok {
		return false
	}
	go func() {
		defer release()
		d.handler(ctx, update)
	}()
	return true
}

// Reserve acquires one dispatch slot and accounts for it in Wait. Internal
// synchronous update sources use it to share the same global concurrency
// bound as queued polling and webhook updates.
func (d *Dispatcher) Reserve(ctx context.Context, wait bool) (release func(), ok bool) {
	if d == nil {
		return nil, false
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return nil, false
	}
	if wait {
		select {
		case d.slots <- struct{}{}:
		case <-ctx.Done():
			return nil, false
		}
	} else {
		select {
		case d.slots <- struct{}{}:
		default:
			return nil, false
		}
	}

	d.mu.Lock()
	d.active++
	d.mu.Unlock()
	return d.done, true
}

// Wait blocks until the dispatcher has no active handlers.
func (d *Dispatcher) Wait() {
	if d == nil {
		return
	}
	d.mu.Lock()
	for d.active != 0 {
		d.drained.Wait()
	}
	d.mu.Unlock()
}

func (d *Dispatcher) done() {
	<-d.slots
	d.mu.Lock()
	d.active--
	if d.active == 0 {
		d.drained.Broadcast()
	}
	d.mu.Unlock()
}
