//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/sidwiskers/hermes"
	"github.com/sidwiskers/hermes/api"
	"github.com/sidwiskers/hermes/observe"
)

// TestLiveSoak is intentionally disabled unless HERMES_TEST_SOAK_DURATION is
// set. It uses the same dedicated credentials and endpoint selection as
// live_test.go and archives a redaction-safe report in go test output.
func TestLiveSoak(t *testing.T) {
	durationText := os.Getenv("HERMES_TEST_SOAK_DURATION")
	if durationText == "" {
		t.Skip("set HERMES_TEST_SOAK_DURATION to enable the live soak")
	}
	duration, err := time.ParseDuration(durationText)
	if err != nil || duration < time.Minute {
		t.Fatalf("HERMES_TEST_SOAK_DURATION must be a duration of at least one minute")
	}
	token := os.Getenv("HERMES_TEST_BOT_TOKEN")
	if token == "" {
		t.Fatal("HERMES_TEST_BOT_TOKEN is required")
	}
	maximum := 64
	if text := os.Getenv("HERMES_TEST_SOAK_CONCURRENCY"); text != "" {
		value, parseErr := strconv.Atoi(text)
		if parseErr != nil || value <= 0 {
			t.Fatal("HERMES_TEST_SOAK_CONCURRENCY must be a positive integer")
		}
		maximum = value
	}

	metrics := new(observe.Metrics)
	errorsSeen := new(errorObserver)
	observer := multiObserver{metrics, errorsSeen}
	options := liveEndpointOptions(t)
	options = append(options,
		hermes.WithMaxConcurrentUpdates(maximum),
		hermes.WithAPIObserver(observer),
		hermes.WithErrorHandler(func(_ *hermes.Context, err error) { errorsSeen.add(err) }),
	)
	bot := hermes.New(token, options...)
	bot.Use(observe.Middleware(metrics))
	bot.OnUpdate(func(*hermes.Context) error { return nil })

	runtime.GC()
	startGoroutines := runtime.NumGoroutine()
	var startMemory runtime.MemStats
	runtime.ReadMemStats(&startMemory)
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	started := time.Now()
	err = bot.Run(ctx)
	cancel()
	bot.Wait()
	elapsed := time.Since(started)
	if err != nil {
		t.Fatalf("live soak: %v", err)
	}
	runtime.GC()
	var endMemory runtime.MemStats
	runtime.ReadMemStats(&endMemory)
	report := map[string]any{
		"duration":            elapsed.String(),
		"go_version":          runtime.Version(),
		"maximum_concurrency": maximum,
		"metrics":             metrics.Snapshot(),
		"non_cancel_errors":   errorsSeen.count(),
		"start_heap_bytes":    startMemory.HeapAlloc,
		"end_heap_bytes":      endMemory.HeapAlloc,
		"heap_delta_bytes":    int64(endMemory.HeapAlloc) - int64(startMemory.HeapAlloc),
		"start_goroutines":    startGoroutines,
		"end_goroutines":      runtime.NumGoroutine(),
	}
	data, _ := json.Marshal(report)
	t.Logf("live soak report: %s", data)
	if errorsSeen.count() != 0 {
		t.Fatalf("live soak observed %d non-cancellation errors", errorsSeen.count())
	}
}

type errorObserver struct {
	mu     sync.Mutex
	errors []error
}

func (o *errorObserver) add(err error) {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return
	}
	o.mu.Lock()
	o.errors = append(o.errors, err)
	o.mu.Unlock()
}

func (o *errorObserver) count() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return len(o.errors)
}

func (o *errorObserver) StartCall(ctx context.Context, _ api.CallEvent) context.Context { return ctx }

func (o *errorObserver) FinishCall(_ context.Context, _ api.CallEvent, result api.CallResult) {
	o.add(result.Err)
}

type multiObserver []api.Observer

func (observers multiObserver) StartCall(ctx context.Context, event api.CallEvent) context.Context {
	for _, observer := range observers {
		if observer != nil {
			if next := observer.StartCall(ctx, event); next != nil {
				ctx = next
			}
		}
	}
	return ctx
}

func (observers multiObserver) FinishCall(ctx context.Context, event api.CallEvent, result api.CallResult) {
	for index := len(observers) - 1; index >= 0; index-- {
		if observers[index] != nil {
			observers[index].FinishCall(ctx, event, result)
		}
	}
}
