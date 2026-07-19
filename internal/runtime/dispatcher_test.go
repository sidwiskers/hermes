package runtime

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	telegram "github.com/sidwiskers/hermes/types"
)

func TestDispatcherBackpressureAndDrain(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	release := make(chan struct{})
	var handled atomic.Int32
	dispatcher := NewDispatcher(1, func(context.Context, *telegram.Update) {
		if handled.Add(1) == 1 {
			close(started)
			<-release
		}
	})

	if !dispatcher.Queue(context.Background(), &telegram.Update{UpdateID: 1}, false) {
		t.Fatal("first update was rejected")
	}
	<-started
	if dispatcher.Queue(context.Background(), &telegram.Update{UpdateID: 2}, false) {
		t.Fatal("full dispatcher accepted non-blocking update")
	}
	close(release)
	dispatcher.Wait()
	if handled.Load() != 1 {
		t.Fatalf("handled = %d", handled.Load())
	}
}

func TestDispatcherBlockingQueueHonorsCancellation(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	dispatcher := NewDispatcher(1, func(context.Context, *telegram.Update) { <-release })
	if !dispatcher.Queue(context.Background(), &telegram.Update{UpdateID: 1}, false) {
		t.Fatal("first update was rejected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if dispatcher.Queue(ctx, &telegram.Update{UpdateID: 2}, true) {
		t.Fatal("queue succeeded after context cancellation")
	}
	close(release)
	dispatcher.Wait()
}

func TestDispatcherNilUpdateIsNoop(t *testing.T) {
	t.Parallel()
	if !NewDispatcher(1, func(context.Context, *telegram.Update) {}).Queue(context.Background(), nil, false) {
		t.Fatal("nil update should be accepted as a no-op")
	}
}

func TestDispatcherRejectsAlreadyCanceledContext(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var handled atomic.Bool
	dispatcher := NewDispatcher(1, func(context.Context, *telegram.Update) {
		handled.Store(true)
	})
	if dispatcher.Queue(ctx, &telegram.Update{UpdateID: 1}, false) {
		t.Fatal("non-blocking queue accepted a canceled context")
	}
	if release, ok := dispatcher.Reserve(ctx, true); ok || release != nil {
		t.Fatal("blocking reserve accepted a canceled context")
	}
	dispatcher.Wait()
	if handled.Load() {
		t.Fatal("canceled update was handled")
	}
}

func TestDispatcherWaitIncludesConcurrentQueue(t *testing.T) {
	t.Parallel()

	firstStarted := make(chan struct{})
	secondStarted := make(chan struct{})
	release := make(chan struct{})
	dispatcher := NewDispatcher(2, func(_ context.Context, update *telegram.Update) {
		if update.UpdateID == 1 {
			close(firstStarted)
		} else {
			close(secondStarted)
		}
		<-release
	})
	if !dispatcher.Queue(context.Background(), &telegram.Update{UpdateID: 1}, false) {
		t.Fatal("first update was rejected")
	}
	<-firstStarted

	drained := make(chan struct{})
	go func() {
		dispatcher.Wait()
		close(drained)
	}()
	if !dispatcher.Queue(context.Background(), &telegram.Update{UpdateID: 2}, false) {
		t.Fatal("second update was rejected")
	}
	<-secondStarted
	select {
	case <-drained:
		t.Fatal("Wait returned while updates were active")
	default:
	}
	close(release)
	select {
	case <-drained:
	case <-time.After(time.Second):
		t.Fatal("Wait did not return after drain")
	}
}
