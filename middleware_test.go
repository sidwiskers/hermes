package hermes

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRecoverMiddleware(t *testing.T) {
	t.Parallel()

	bot := New("TOKEN")
	bot.Use(Recover())
	bot.OnUpdate(func(*Context) error {
		panic("boom")
	})

	err := bot.Handle(context.Background(), &Update{})
	var panicErr *PanicError
	if !errors.As(err, &panicErr) {
		t.Fatalf("expected PanicError, got %T: %v", err, err)
	}
	if len(panicErr.Stack) == 0 {
		t.Fatal("missing panic stack")
	}
}

func TestTimeoutMiddleware(t *testing.T) {
	t.Parallel()

	bot := New("TOKEN")
	bot.Use(Timeout(time.Millisecond))
	bot.OnUpdate(func(c *Context) error {
		<-c.Done()
		return c.Err()
	})

	err := bot.Handle(context.Background(), &Update{})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline, got %v", err)
	}
}
