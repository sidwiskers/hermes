package framework

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestRecoverMiddleware(t *testing.T) {
	t.Parallel()
	handler := Recover()(func(*Context) error { panic("boom") })
	err := handler(&Context{Context: context.Background()})
	var panicErr *PanicError
	if !errors.As(err, &panicErr) || panicErr.Value != "boom" || len(panicErr.Stack) == 0 {
		t.Fatalf("panic error = %#v", err)
	}
}

func TestRecoverWithReports(t *testing.T) {
	t.Parallel()
	reported := false
	handler := RecoverWith(func(_ *Context, err *PanicError) {
		reported = err.Value == 7
	})(func(*Context) error { panic(7) })
	if err := handler(&Context{Context: context.Background()}); err == nil || !reported {
		t.Fatalf("err=%v reported=%v", err, reported)
	}
}

func TestTimeoutMiddleware(t *testing.T) {
	t.Parallel()
	handler := Timeout(5 * time.Millisecond)(func(c *Context) error {
		<-c.Done()
		return c.Err()
	})
	err := handler(&Context{Context: context.Background()})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout error = %v", err)
	}
}

func TestLoggerMiddleware(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ctx := filterContext()
	handler := Logger(logger)(func(*Context) error { return errors.New("failed") })
	if err := handler(ctx); err == nil {
		t.Fatal("handler error was swallowed")
	}
	text := output.String()
	for _, expected := range []string{"telegram update failed", "update_id=", "user_id=11", "chat_id=22"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("log %q missing %q", text, expected)
		}
	}
}
