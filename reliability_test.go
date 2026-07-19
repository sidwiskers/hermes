package hermes

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/sidwiskers/hermes/framework"
)

type reliabilityRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn reliabilityRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestAsyncHandlerPanicIsReportedAndDispatcherContinues(t *testing.T) {
	t.Parallel()

	reported := make(chan error, 1)
	handled := make(chan struct{}, 1)
	bot := New("TOKEN",
		WithMaxConcurrentUpdates(1),
		WithErrorHandler(func(_ *Context, err error) { reported <- err }),
	)
	bot.OnUpdate(func(c *Context) error {
		if c.Update.UpdateID == 1 {
			panic("boom")
		}
		handled <- struct{}{}
		return nil
	})

	if !bot.queue(context.Background(), &Update{UpdateID: 1}, true) {
		t.Fatal("panic update was not queued")
	}
	if !bot.queue(context.Background(), &Update{UpdateID: 2}, true) {
		t.Fatal("following update was not queued")
	}
	bot.Wait()

	select {
	case err := <-reported:
		var panicErr *framework.PanicError
		if !errors.As(err, &panicErr) || panicErr.Value != "boom" || len(panicErr.Stack) == 0 {
			t.Fatalf("reported error = %#v", err)
		}
	default:
		t.Fatal("panic was not reported")
	}
	select {
	case <-handled:
	default:
		t.Fatal("dispatcher did not continue after panic")
	}
}

func TestNilBotOperationsReturnError(t *testing.T) {
	t.Parallel()

	var bot *Bot
	if err := bot.Handle(context.Background(), &Update{}); !errors.Is(err, ErrClientRequired) {
		t.Fatalf("Handle error = %v", err)
	}
	if err := bot.Poll(context.Background(), PollOptions{}); !errors.Is(err, ErrClientRequired) {
		t.Fatalf("Poll error = %v", err)
	}
	if err := bot.Run(context.Background()); !errors.Is(err, ErrClientRequired) {
		t.Fatalf("Run error = %v", err)
	}
	if err := bot.ServeWebhook(context.Background(), "127.0.0.1:0", "/hook", WebhookOptions{}); !errors.Is(err, ErrClientRequired) {
		t.Fatalf("ServeWebhook error = %v", err)
	}
	if _, err := Call[bool](context.Background(), bot, "getMe", nil); !errors.Is(err, ErrClientRequired) {
		t.Fatalf("Call error = %v", err)
	}
}

func TestEnsureUsernameIsConcurrentSafe(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	client := &http.Client{Transport: reliabilityRoundTripFunc(func(*http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(
				`{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Hermes","username":"Hermes_Bot"}}`,
			)),
		}, nil
	})}
	bot := New("TOKEN", WithHTTPClient(client))
	const workers = 32
	errorsSeen := make(chan error, workers)
	var group sync.WaitGroup
	group.Add(workers)
	for range workers {
		go func() {
			defer group.Done()
			errorsSeen <- bot.ensureUsername(context.Background())
		}()
	}
	group.Wait()
	close(errorsSeen)
	for err := range errorsSeen {
		if err != nil {
			t.Fatal(err)
		}
	}
	if calls.Load() != 1 || bot.loadUsername() != "hermes_bot" {
		t.Fatalf("calls=%d username=%q", calls.Load(), bot.loadUsername())
	}
}

func TestAsyncErrorHandlerPanicIsContained(t *testing.T) {
	bot := New("TOKEN",
		WithMaxConcurrentUpdates(1),
		WithErrorHandler(func(*Context, error) { panic("reporter panic") }),
	)
	var handled atomic.Int32
	bot.OnUpdate(func(*Context) error {
		handled.Add(1)
		return errors.New("handler error")
	})
	if !bot.queue(context.Background(), &Update{UpdateID: 1}, true) ||
		!bot.queue(context.Background(), &Update{UpdateID: 2}, true) {
		t.Fatal("updates were not queued")
	}
	bot.Wait()
	if handled.Load() != 2 {
		t.Fatalf("handled = %d", handled.Load())
	}
}
