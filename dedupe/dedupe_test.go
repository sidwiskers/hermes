package dedupe

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sidwiskers/hermes/framework"
	"github.com/sidwiskers/hermes/types"
)

func TestManagerSuppressesDuplicate(t *testing.T) {
	store := NewMemory()
	manager := New(store, WithNamespace("bot"))
	var handled atomic.Int32
	handler := manager.Middleware()(func(*framework.Context) error {
		handled.Add(1)
		return nil
	})
	ctx := testContext(10)
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if err := handler(ctx); err != nil {
		t.Fatal(err)
	}
	if handled.Load() != 1 || store.Len() != 1 {
		t.Fatalf("handled=%d claims=%d", handled.Load(), store.Len())
	}
}

func TestManagerReleasesErrorAndPanic(t *testing.T) {
	store := NewMemory()
	manager := New(store)
	wantErr := errors.New("failed")
	handler := manager.Middleware()(func(*framework.Context) error { return wantErr })
	if err := handler(testContext(11)); !errors.Is(err, wantErr) || store.Len() != 0 {
		t.Fatalf("error=%v claims=%d", err, store.Len())
	}
	panics := manager.Middleware()(func(*framework.Context) error { panic("boom") })
	func() {
		defer func() {
			if recover() != "boom" {
				t.Fatal("panic not preserved")
			}
		}()
		_ = panics(testContext(12))
	}()
	if store.Len() != 0 {
		t.Fatalf("panic claim retained: %d", store.Len())
	}
}

func TestManagerConcurrentClaimIsAtomic(t *testing.T) {
	store := NewMemory()
	manager := New(store)
	var handled atomic.Int32
	release := make(chan struct{})
	handler := manager.Middleware()(func(*framework.Context) error {
		handled.Add(1)
		<-release
		return nil
	})
	var group sync.WaitGroup
	for range 32 {
		group.Add(1)
		go func() {
			defer group.Done()
			_ = handler(testContext(13))
		}()
	}
	deadline := time.After(time.Second)
	for handled.Load() == 0 {
		select {
		case <-deadline:
			close(release)
			group.Wait()
			t.Fatal("claimed handler did not start")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	close(release)
	group.Wait()
	if handled.Load() != 1 {
		t.Fatalf("handled=%d", handled.Load())
	}
}

func TestMemoryExpiryCapacityAndRelease(t *testing.T) {
	now := time.Unix(100, 0)
	store := NewMemory(MemoryConfig{MaxEntries: 1, Shards: 3})
	store.now = func() time.Time { return now }
	ctx := context.Background()
	first := Key{UpdateID: 1}
	if claimed, err := store.Claim(ctx, first, time.Minute); err != nil || !claimed {
		t.Fatalf("claim=%v err=%v", claimed, err)
	}
	if claimed, err := store.Claim(ctx, first, time.Minute); err != nil || claimed {
		t.Fatalf("duplicate claim=%v err=%v", claimed, err)
	}
	if _, err := store.Claim(ctx, Key{UpdateID: 2}, time.Minute); !errors.Is(err, ErrCapacity) {
		t.Fatalf("capacity error=%v", err)
	}
	now = now.Add(time.Minute)
	if removed := store.Sweep(); removed != 1 || store.Len() != 0 {
		t.Fatalf("removed=%d len=%d", removed, store.Len())
	}
	if claimed, err := store.Claim(ctx, first, time.Minute); err != nil || !claimed {
		t.Fatalf("reclaim=%v err=%v", claimed, err)
	}
	if err := store.Release(ctx, first); err != nil || store.Len() != 0 {
		t.Fatalf("release err=%v len=%d", err, store.Len())
	}
}

func TestMemoryConcurrentBoundNeverExceeded(t *testing.T) {
	store := NewMemory(MemoryConfig{MaxEntries: 8})
	var group sync.WaitGroup
	for index := range 64 {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := store.Claim(
				context.Background(),
				Key{UpdateID: int64(index + 1)},
				time.Minute,
			)
			if err != nil && !errors.Is(err, ErrCapacity) {
				t.Errorf("claim: %v", err)
			}
		}()
	}
	group.Wait()
	if store.Len() != 8 {
		t.Fatalf("len=%d, want 8", store.Len())
	}
	if retained := retainedMemoryEntries(store); retained != store.Len() {
		t.Fatalf("retained entries=%d, counter=%d", retained, store.Len())
	}
}

func TestMemoryShardMapsAreLazyAndReleased(t *testing.T) {
	store := NewMemory()
	assertAllocatedShardMaps(t, store, 0)

	key := Key{UpdateID: 1}
	if claimed, err := store.Claim(context.Background(), key, time.Minute); err != nil || !claimed {
		t.Fatalf("claim=%v err=%v", claimed, err)
	}
	assertAllocatedShardMaps(t, store, 1)

	if err := store.Release(context.Background(), key); err != nil {
		t.Fatal(err)
	}
	assertAllocatedShardMaps(t, store, 0)
}

func TestMemoryReleasesShardMapsAfterChurn(t *testing.T) {
	store := NewMemory(MemoryConfig{MaxEntries: 0})
	ctx := context.Background()
	const entries = 4_096
	for index := range entries {
		key := Key{Namespace: "churn", UpdateID: int64(index + 1)}
		if claimed, err := store.Claim(ctx, key, time.Hour); err != nil || !claimed {
			t.Fatalf("claim=%v err=%v", claimed, err)
		}
	}
	if retained := retainedMemoryEntries(store); retained != entries {
		t.Fatalf("retained entries=%d, want %d", retained, entries)
	}
	for index := range entries {
		key := Key{Namespace: "churn", UpdateID: int64(index + 1)}
		if err := store.Release(ctx, key); err != nil {
			t.Fatal(err)
		}
	}
	if store.Len() != 0 {
		t.Fatalf("len=%d, want 0", store.Len())
	}
	assertAllocatedShardMaps(t, store, 0)
}

func TestMemoryDuplicateClaimIsAllocationFree(t *testing.T) {
	store := NewMemory()
	ctx := context.Background()
	key := Key{Namespace: "steady", UpdateID: 1}
	if claimed, err := store.Claim(ctx, key, time.Hour); err != nil || !claimed {
		t.Fatalf("claim=%v err=%v", claimed, err)
	}
	allocations := testing.AllocsPerRun(1_000, func() {
		if claimed, err := store.Claim(ctx, key, time.Hour); err != nil || claimed {
			panic("duplicate claim mismatch")
		}
	})
	if allocations != 0 {
		t.Fatalf("duplicate-claim allocations=%v", allocations)
	}
}

func assertAllocatedShardMaps(t *testing.T, store *Memory, want int) {
	t.Helper()
	allocated := 0
	for index := range store.shards {
		if store.shards[index].claims != nil {
			allocated++
		}
	}
	if allocated != want {
		t.Fatalf("allocated shard maps=%d, want %d", allocated, want)
	}
}

func retainedMemoryEntries(store *Memory) int {
	retained := 0
	for index := range store.shards {
		retained += len(store.shards[index].claims)
	}
	return retained
}

func BenchmarkMemoryConstruction(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		store := NewMemory()
		runtime.KeepAlive(store)
	}
}

func TestDuplicateCallbackAndSkippedUpdate(t *testing.T) {
	store := NewMemory()
	var duplicates atomic.Int32
	manager := New(store, WithDuplicate(func(*framework.Context) error {
		duplicates.Add(1)
		return nil
	}))
	var handled atomic.Int32
	handler := manager.Middleware()(func(*framework.Context) error {
		handled.Add(1)
		return nil
	})
	if err := handler(testContext(14)); err != nil {
		t.Fatal(err)
	}
	if err := handler(testContext(14)); err != nil {
		t.Fatal(err)
	}
	if err := handler(testContext(0)); err != nil {
		t.Fatal(err)
	}
	if handled.Load() != 2 || duplicates.Load() != 1 {
		t.Fatalf("handled=%d duplicates=%d", handled.Load(), duplicates.Load())
	}
}

func testContext(updateID int64) *framework.Context {
	return framework.NewContext(context.Background(), nil, &types.Update{UpdateID: updateID}, "")
}
