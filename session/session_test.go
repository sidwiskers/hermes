package session

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sidwiskers/hermes/framework"
	"github.com/sidwiskers/hermes/types"
)

func TestManagerPersistsSuccessfulMutation(t *testing.T) {
	store := NewMemory[int](0)
	manager := New[int](store, ByChatUser, WithNamespace("counter"))
	handler := manager.Middleware()(func(c *framework.Context) error {
		value, exists, err := manager.Get(c)
		if err != nil || exists || value != 0 {
			t.Fatalf("unexpected initial session: value=%d exists=%v err=%v", value, exists, err)
		}
		return manager.Set(c, 42)
	})

	if err := handler(testContext(11, 22)); err != nil {
		t.Fatal(err)
	}
	value, exists, err := store.Load(context.Background(), Key{Namespace: "counter", ChatID: 11, UserID: 22})
	if err != nil || !exists || value != 42 {
		t.Fatalf("stored value=%d exists=%v err=%v", value, exists, err)
	}
}

func TestManagerRollsBackHandlerError(t *testing.T) {
	store := NewMemory[int](0)
	key := Key{ChatID: 11, UserID: 22}
	if err := store.Save(context.Background(), key, 7); err != nil {
		t.Fatal(err)
	}
	manager := New[int](store, ByChatUser)
	wantErr := errors.New("failed")
	handler := manager.Middleware()(func(c *framework.Context) error {
		if err := manager.Set(c, 8); err != nil {
			return err
		}
		return wantErr
	})
	if err := handler(testContext(11, 22)); !errors.Is(err, wantErr) {
		t.Fatalf("error = %v", err)
	}
	value, _, _ := store.Load(context.Background(), key)
	if value != 7 {
		t.Fatalf("value committed on error: %d", value)
	}
}

func TestManagerCommitOnErrorAndDelete(t *testing.T) {
	store := NewMemory[int](0)
	manager := New[int](store, ByUser, WithCommitOnError(true))
	wantErr := errors.New("failed")
	set := manager.Middleware()(func(c *framework.Context) error {
		if err := manager.Set(c, 9); err != nil {
			return err
		}
		return wantErr
	})
	if err := set(testContext(0, 22)); !errors.Is(err, wantErr) {
		t.Fatal(err)
	}
	key := Key{UserID: 22}
	if value, exists, _ := store.Load(context.Background(), key); !exists || value != 9 {
		t.Fatalf("commit-on-error value=%d exists=%v", value, exists)
	}
	remove := manager.Middleware()(func(c *framework.Context) error { return manager.Delete(c) })
	if err := remove(testContext(0, 22)); err != nil {
		t.Fatal(err)
	}
	if _, exists, _ := store.Load(context.Background(), key); exists {
		t.Fatal("deleted session still exists")
	}
}

func TestManagerSerializesSameKey(t *testing.T) {
	store := NewMemory[int](0)
	manager := New[int](store, ByUser)
	var active atomic.Int32
	var overlapped atomic.Bool
	handler := manager.Middleware()(func(c *framework.Context) error {
		if active.Add(1) != 1 {
			overlapped.Store(true)
		}
		time.Sleep(time.Millisecond)
		err := manager.Update(c, func(value *int) error {
			*value++
			return nil
		})
		active.Add(-1)
		return err
	})

	var group sync.WaitGroup
	for range 16 {
		group.Add(1)
		go func() {
			defer group.Done()
			if err := handler(testContext(0, 22)); err != nil {
				t.Errorf("handler: %v", err)
			}
		}()
	}
	group.Wait()
	if overlapped.Load() {
		t.Fatal("same session key executed concurrently")
	}
	value, _, _ := store.Load(context.Background(), Key{UserID: 22})
	if value != 16 {
		t.Fatalf("value=%d, want 16", value)
	}
}

func TestManagerDifferentKeysCanOverlap(t *testing.T) {
	store := NewMemory[int](0)
	manager := New[int](store, func(c *framework.Context) (Key, bool) {
		return Key{UserID: c.Sender().ID}, true
	})
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	handler := manager.Middleware()(func(*framework.Context) error {
		started <- struct{}{}
		<-release
		return nil
	})

	for _, userID := range []int64{1, 2} {
		go func() { _ = handler(testContext(0, userID)) }()
	}
	for range 2 {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("different session keys were unnecessarily serialized")
		}
	}
	close(release)
}

func TestMemoryExpiryCapacityAndSweep(t *testing.T) {
	now := time.Unix(100, 0)
	store := NewMemoryWithConfig[int](MemoryConfig{TTL: time.Minute, MaxEntries: 2, Shards: 3})
	store.now = func() time.Time { return now }
	ctx := context.Background()
	first := Key{UserID: 1}
	second := Key{UserID: 2}
	if err := store.Save(ctx, first, 1); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(ctx, second, 2); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(ctx, Key{UserID: 3}, 3); !errors.Is(err, ErrCapacity) {
		t.Fatalf("capacity error = %v", err)
	}
	now = now.Add(time.Minute)
	if value, exists, err := store.Load(ctx, first); err != nil || exists || value != 0 {
		t.Fatalf("expired load value=%d exists=%v err=%v", value, exists, err)
	}
	if removed := store.Sweep(); removed != 1 {
		t.Fatalf("removed=%d, want 1", removed)
	}
	if store.Len() != 0 {
		t.Fatalf("len=%d", store.Len())
	}
}

func TestMemoryHonorsCancellation(t *testing.T) {
	store := NewMemory[int](0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := store.Save(ctx, Key{UserID: 1}, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("save error=%v", err)
	}
}

func TestSessionUnavailableAndSkippedKey(t *testing.T) {
	manager := New[int](NewMemory[int](0), ByUser)
	if _, _, err := manager.Get(testContext(1, 1)); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("get error=%v", err)
	}
	called := false
	handler := New[int](NewMemory[int](0), func(*framework.Context) (Key, bool) {
		return Key{}, false
	}).Middleware()(func(*framework.Context) error {
		called = true
		return nil
	})
	if err := handler(testContext(0, 0)); err != nil || !called {
		t.Fatalf("skipped handler called=%v err=%v", called, err)
	}
}

func testContext(chatID, userID int64) *framework.Context {
	update := &types.Update{Message: &types.Message{
		Chat: types.Chat{ID: chatID},
		From: &types.User{ID: userID},
	}}
	return framework.NewContext(context.Background(), nil, update, "")
}
