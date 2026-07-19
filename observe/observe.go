package observe

import (
	"context"
	"time"

	"github.com/sidwiskers/hermes/framework"
)

// UpdateEvent contains bounded-cardinality routing metadata for one update.
type UpdateEvent struct {
	UpdateID   int64
	UpdateType framework.UpdateType
	UserID     int64
	ChatID     int64
	Command    string
}

// UpdateResult describes a completed handler.
type UpdateResult struct {
	Duration time.Duration
	Err      error
	Panicked bool
}

// UpdateObserver receives handler lifecycle events. StartUpdate may return a
// derived context for trace propagation. Observer panics are contained.
type UpdateObserver interface {
	StartUpdate(context.Context, UpdateEvent) context.Context
	FinishUpdate(context.Context, UpdateEvent, UpdateResult)
}

// Hooks adapts functions into an UpdateObserver.
type Hooks struct {
	Start  func(context.Context, UpdateEvent) context.Context
	Finish func(context.Context, UpdateEvent, UpdateResult)
}

// StartUpdate implements UpdateObserver.
func (h Hooks) StartUpdate(ctx context.Context, event UpdateEvent) context.Context {
	if h.Start == nil {
		return ctx
	}
	return h.Start(ctx, event)
}

// FinishUpdate implements UpdateObserver.
func (h Hooks) FinishUpdate(ctx context.Context, event UpdateEvent, result UpdateResult) {
	if h.Finish != nil {
		h.Finish(ctx, event, result)
	}
}

// Middleware observes one downstream handler invocation. It preserves panics
// for Hermes recovery while marking them in UpdateResult.
func Middleware(observer UpdateObserver) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(c *framework.Context) (err error) {
			if next == nil {
				return nil
			}
			if observer == nil {
				return next(c)
			}
			event := eventFromContext(c)
			ctx := context.Background()
			if c != nil && c.Context != nil {
				ctx = c.Context
			}
			started := time.Now()
			ctx = safeStart(observer, ctx, event)
			cloned := c
			if c != nil {
				value := *c
				value.Context = ctx
				cloned = &value
			}
			defer func() {
				if value := recover(); value != nil {
					safeFinish(observer, ctx, event, UpdateResult{
						Duration: time.Since(started),
						Panicked: true,
					})
					panic(value)
				}
				safeFinish(observer, ctx, event, UpdateResult{
					Duration: time.Since(started),
					Err:      err,
				})
			}()
			return next(cloned)
		}
	}
}

func eventFromContext(c *framework.Context) UpdateEvent {
	if c == nil {
		return UpdateEvent{}
	}
	event := UpdateEvent{UpdateType: c.Type(), Command: c.Command()}
	if c.Update != nil {
		event.UpdateID = c.Update.UpdateID
	}
	if user := c.Sender(); user != nil {
		event.UserID = user.ID
	}
	if chatID, ok := c.ChatID(); ok {
		event.ChatID = chatID
	}
	return event
}

func safeStart(observer UpdateObserver, ctx context.Context, event UpdateEvent) (result context.Context) {
	result = ctx
	defer func() { _ = recover() }()
	if observed := observer.StartUpdate(ctx, event); observed != nil {
		result = observed
	}
	return result
}

func safeFinish(observer UpdateObserver, ctx context.Context, event UpdateEvent, result UpdateResult) {
	defer func() { _ = recover() }()
	observer.FinishUpdate(ctx, event, result)
}
