package ratelimit

import (
	"context"
	"errors"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sidwiskers/hermes/framework"
	"github.com/sidwiskers/hermes/types"
)

func TestLimiterTokenRefill(t *testing.T) {
	limiter, err := New(2, time.Second, WithBurst(2))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(100, 0)
	limiter.now = func() time.Time { return now }
	key := Key{UserID: 1}

	first, _ := limiter.Allow(key)
	second, _ := limiter.Allow(key)
	third, _ := limiter.Allow(key)
	if !first.Allowed || first.Remaining != 1 || !second.Allowed || third.Allowed {
		t.Fatalf("decisions: %+v %+v %+v", first, second, third)
	}
	if third.RetryAfter != 500*time.Millisecond {
		t.Fatalf("retry=%s", third.RetryAfter)
	}
	now = now.Add(500 * time.Millisecond)
	refilled, _ := limiter.Allow(key)
	if !refilled.Allowed || refilled.Remaining != 0 {
		t.Fatalf("refilled=%+v", refilled)
	}
}

func TestMiddlewareRejectsAndCustomHandlerConsumes(t *testing.T) {
	var rejected atomic.Int32
	limiter, err := New(1, time.Hour, WithRejected(func(_ *framework.Context, decision Decision) error {
		if decision.RetryAfter <= 0 {
			t.Fatal("missing retry delay")
		}
		rejected.Add(1)
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	var handled atomic.Int32
	handler := limiter.Middleware()(func(*framework.Context) error {
		handled.Add(1)
		return nil
	})
	ctx := testContext(1, 2)
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if handled.Load() != 1 || rejected.Load() != 1 {
		t.Fatalf("handled=%d rejected=%d", handled.Load(), rejected.Load())
	}
}

func TestMiddlewareReturnsTypedLimitError(t *testing.T) {
	limiter, _ := New(1, time.Hour)
	handler := limiter.Middleware()(func(*framework.Context) error { return nil })
	ctx := testContext(1, 2)
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	err := handler(ctx)
	var limitErr *LimitError
	if !errors.Is(err, ErrLimited) || !errors.As(err, &limitErr) || limitErr.Key.UserID != 2 {
		t.Fatalf("error=%v", err)
	}
}

func TestCapacityExistingKeysAndSweep(t *testing.T) {
	limiter, err := New(1, time.Second, WithMaxKeys(1), WithIdleTTL(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(100, 0)
	limiter.now = func() time.Time { return now }
	first := Key{UserID: 1}
	if _, err := limiter.Allow(first); err != nil {
		t.Fatal(err)
	}
	if _, err := limiter.Allow(first); err != nil {
		t.Fatalf("existing key rejected at capacity: %v", err)
	}
	if _, err := limiter.Allow(Key{UserID: 2}); !errors.Is(err, ErrCapacity) {
		t.Fatalf("capacity error=%v", err)
	}
	now = now.Add(time.Minute)
	if removed := limiter.Sweep(); removed != 1 || limiter.Len() != 0 {
		t.Fatalf("removed=%d len=%d", removed, limiter.Len())
	}
	if _, err := limiter.Allow(Key{UserID: 2}); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentBoundNeverExceeded(t *testing.T) {
	limiter, _ := New(1, time.Second, WithMaxKeys(8))
	var group sync.WaitGroup
	for index := range 64 {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := limiter.Allow(Key{UserID: int64(index + 1)})
			if err != nil && !errors.Is(err, ErrCapacity) {
				t.Errorf("allow: %v", err)
			}
		}()
	}
	group.Wait()
	if limiter.Len() != 8 {
		t.Fatalf("len=%d, want 8", limiter.Len())
	}
	if retained := retainedLimiterEntries(limiter); retained != limiter.Len() {
		t.Fatalf("retained entries=%d, counter=%d", retained, limiter.Len())
	}
}

func TestLimiterShardMapsAreLazyAndReleased(t *testing.T) {
	limiter, err := New(1, time.Second, WithIdleTTL(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	assertAllocatedShardMaps(t, limiter, 0)

	now := time.Unix(100, 0)
	limiter.now = func() time.Time { return now }
	if _, err := limiter.Allow(Key{UserID: 1}); err != nil {
		t.Fatal(err)
	}
	assertAllocatedShardMaps(t, limiter, 1)

	now = now.Add(time.Minute)
	if removed := limiter.Sweep(); removed != 1 {
		t.Fatalf("removed=%d, want 1", removed)
	}
	assertAllocatedShardMaps(t, limiter, 0)
}

func TestLimiterReleasesShardMapsAfterChurn(t *testing.T) {
	limiter, err := New(1, time.Second, WithMaxKeys(0), WithIdleTTL(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(100, 0)
	limiter.now = func() time.Time { return now }
	const entries = 4_096
	for index := range entries {
		if _, err := limiter.Allow(Key{
			Namespace: "churn",
			UserID:    int64(index + 1),
		}); err != nil {
			t.Fatal(err)
		}
	}
	if retained := retainedLimiterEntries(limiter); retained != entries {
		t.Fatalf("retained entries=%d, want %d", retained, entries)
	}
	now = now.Add(time.Minute)
	if removed := limiter.Sweep(); removed != entries {
		t.Fatalf("removed=%d, want %d", removed, entries)
	}
	if limiter.Len() != 0 {
		t.Fatalf("len=%d, want 0", limiter.Len())
	}
	assertAllocatedShardMaps(t, limiter, 0)
}

func TestLimiterExistingKeyIsAllocationFree(t *testing.T) {
	limiter, err := New(1, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(100, 0)
	limiter.now = func() time.Time { return now }
	key := Key{Namespace: "steady", UserID: 1}
	if _, err := limiter.Allow(key); err != nil {
		t.Fatal(err)
	}
	allocations := testing.AllocsPerRun(1_000, func() {
		if _, err := limiter.Allow(key); err != nil {
			panic(err)
		}
	})
	if allocations != 0 {
		t.Fatalf("existing-key allocations=%v", allocations)
	}
}

func assertAllocatedShardMaps(t *testing.T, limiter *Limiter, want int) {
	t.Helper()
	allocated := 0
	for index := range limiter.shards {
		if limiter.shards[index].buckets != nil {
			allocated++
		}
	}
	if allocated != want {
		t.Fatalf("allocated shard maps=%d, want %d", allocated, want)
	}
}

func retainedLimiterEntries(limiter *Limiter) int {
	retained := 0
	for index := range limiter.shards {
		retained += len(limiter.shards[index].buckets)
	}
	return retained
}

func BenchmarkLimiterConstruction(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		limiter, err := New(1, time.Second)
		if err != nil {
			b.Fatal(err)
		}
		runtime.KeepAlive(limiter)
	}
}

func BenchmarkLimiterPopulate1000(b *testing.B) {
	now := time.Unix(100, 0)
	b.ReportAllocs()
	for b.Loop() {
		limiter, err := New(1, time.Second, WithMaxKeys(0))
		if err != nil {
			b.Fatal(err)
		}
		limiter.now = func() time.Time { return now }
		for index := range 1_000 {
			if _, err := limiter.Allow(Key{UserID: int64(index + 1)}); err != nil {
				b.Fatal(err)
			}
		}
		runtime.KeepAlive(limiter)
	}
}

func BenchmarkLimiterDistinctKeysParallel(b *testing.B) {
	limiter, err := New(1, time.Second, WithMaxKeys(0))
	if err != nil {
		b.Fatal(err)
	}
	now := time.Unix(100, 0)
	limiter.now = func() time.Time { return now }
	var sequence atomic.Int64
	b.ReportAllocs()
	b.RunParallel(func(worker *testing.PB) {
		for worker.Next() {
			key := Key{UserID: sequence.Add(1)}
			if _, err := limiter.Allow(key); err != nil {
				b.Error(err)
				return
			}
		}
	})
}

func TestByKeyPoliciesAndInvalidConfig(t *testing.T) {
	ctx := testContext(10, 20)
	if key, ok := ByUser(ctx); !ok || key.UserID != 20 || key.ChatID != 0 {
		t.Fatalf("user key=%+v ok=%v", key, ok)
	}
	if key, ok := ByChat(ctx); !ok || key.ChatID != 10 || key.UserID != 0 {
		t.Fatalf("chat key=%+v ok=%v", key, ok)
	}
	if key, ok := ByChatUser(ctx); !ok || key.ChatID != 10 || key.UserID != 20 {
		t.Fatalf("chat-user key=%+v ok=%v", key, ok)
	}
	if _, err := New(0, time.Second); !errors.Is(err, ErrInvalidLimit) {
		t.Fatalf("invalid error=%v", err)
	}
	if _, err := New(1, time.Second, WithBurst(maxExactBurst+1)); !errors.Is(err, ErrInvalidLimit) {
		t.Fatalf("oversized burst error=%v", err)
	}
}

func TestRetryDurationSaturates(t *testing.T) {
	if got := retryDuration(float64(math.MaxInt64)); got != time.Duration(math.MaxInt64) {
		t.Fatalf("retry duration=%s", got)
	}
	if got := retryDuration(0); got != time.Nanosecond {
		t.Fatalf("minimum retry duration=%s", got)
	}
}

func testContext(chatID, userID int64) *framework.Context {
	update := &types.Update{Message: &types.Message{
		Chat: types.Chat{ID: chatID},
		From: &types.User{ID: userID},
	}}
	return framework.NewContext(context.Background(), nil, update, "")
}
