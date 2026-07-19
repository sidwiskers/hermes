package observe

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sidwiskers/hermes/api"
	"github.com/sidwiskers/hermes/framework"
	"github.com/sidwiskers/hermes/types"
)

type traceKey struct{}

func TestMiddlewarePropagatesContextAndMetadata(t *testing.T) {
	var started UpdateEvent
	var finished UpdateResult
	hooks := Hooks{
		Start: func(ctx context.Context, event UpdateEvent) context.Context {
			started = event
			return context.WithValue(ctx, traceKey{}, "trace")
		},
		Finish: func(_ context.Context, _ UpdateEvent, result UpdateResult) { finished = result },
	}
	wantErr := errors.New("failed")
	handler := Middleware(hooks)(func(c *framework.Context) error {
		if value := c.Value(traceKey{}); value != "trace" {
			t.Fatalf("trace value=%v", value)
		}
		return wantErr
	})
	if err := handler(testContext()); !errors.Is(err, wantErr) {
		t.Fatalf("error=%v", err)
	}
	if started.UpdateID != 9 || started.UserID != 2 || started.ChatID != 1 || started.Command != "start" {
		t.Fatalf("event=%+v", started)
	}
	if !errors.Is(finished.Err, wantErr) || finished.Duration < 0 {
		t.Fatalf("result=%+v", finished)
	}
}

func TestMiddlewareReportsAndPreservesPanic(t *testing.T) {
	var result UpdateResult
	handler := Middleware(Hooks{Finish: func(_ context.Context, _ UpdateEvent, value UpdateResult) {
		result = value
	}})(func(*framework.Context) error { panic("boom") })
	defer func() {
		if recover() != "boom" || !result.Panicked {
			t.Fatalf("panic result=%+v", result)
		}
	}()
	_ = handler(testContext())
}

func TestObserverPanicIsContained(t *testing.T) {
	handler := Middleware(Hooks{
		Start:  func(context.Context, UpdateEvent) context.Context { panic("start") },
		Finish: func(context.Context, UpdateEvent, UpdateResult) { panic("finish") },
	})(func(*framework.Context) error { return nil })
	if err := handler(testContext()); err != nil {
		t.Fatal(err)
	}
}

func TestMetricsTracksUpdatesAndCalls(t *testing.T) {
	metrics := new(Metrics)
	handler := Middleware(metrics)(func(*framework.Context) error { return nil })
	if err := handler(testContext()); err != nil {
		t.Fatal(err)
	}
	wantErr := errors.New("api failed")
	ctx := metrics.StartCall(context.Background(), api.CallEvent{Method: "getMe", Kind: api.CallJSON})
	metrics.FinishCall(ctx, api.CallEvent{}, api.CallResult{Duration: time.Millisecond, Err: wantErr})
	snapshot := metrics.Snapshot()
	if snapshot.UpdatesStarted != 1 || snapshot.UpdatesSucceeded != 1 || snapshot.UpdatesInFlight != 0 {
		t.Fatalf("updates=%+v", snapshot)
	}
	if snapshot.CallsStarted != 1 || snapshot.CallsFailed != 1 || snapshot.CallsInFlight != 0 {
		t.Fatalf("calls=%+v", snapshot)
	}
	if snapshot.AverageCallDuration() != time.Millisecond || snapshot.AverageUpdateDuration() < 0 {
		t.Fatalf("averages update=%s call=%s", snapshot.AverageUpdateDuration(), snapshot.AverageCallDuration())
	}
}

func testContext() *framework.Context {
	update := &types.Update{UpdateID: 9, Message: &types.Message{
		Chat: types.Chat{ID: 1},
		From: &types.User{ID: 2},
		Text: "/start hello",
	}}
	return framework.NewContext(context.Background(), nil, update, "bot")
}
