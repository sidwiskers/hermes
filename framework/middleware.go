package framework

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"
)

// PanicError describes a panic recovered from an update handler.
type PanicError struct {
	Value any
	Stack []byte
}

// Error implements error.
func (e *PanicError) Error() string {
	return fmt.Sprintf("hermes: update handler panicked: %v", e.Value)
}

// Recover converts handler panics into errors without hiding the stack.
func Recover() Middleware {
	return func(next Handler) Handler {
		return func(c *Context) (err error) {
			defer func() {
				if value := recover(); value != nil {
					err = &PanicError{Value: value, Stack: debug.Stack()}
				}
			}()
			return next(c)
		}
	}
}

// Timeout gives each downstream handler a derived deadline.
func Timeout(duration time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(c *Context) error {
			if duration <= 0 {
				return next(c)
			}
			ctx, cancel := context.WithTimeout(c.Context, duration)
			defer cancel()

			cloned := *c
			cloned.Context = ctx
			return next(&cloned)
		}
	}
}

// RecoverWith converts panics into PanicError values and lets the caller
// observe them before they continue through the normal error handler.
func RecoverWith(report func(*Context, *PanicError)) Middleware {
	return func(next Handler) Handler {
		return func(c *Context) (err error) {
			defer func() {
				if value := recover(); value != nil {
					panicErr := &PanicError{Value: value, Stack: debug.Stack()}
					if report != nil {
						report(c, panicErr)
					}
					err = panicErr
				}
			}()
			return next(c)
		}
	}
}
