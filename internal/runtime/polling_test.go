package runtime

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/sidwiskers/hermes/api"
	telegram "github.com/sidwiskers/hermes/types"
)

type scriptedSource struct {
	mu      sync.Mutex
	calls   []api.GetUpdatesParams
	results [][]telegram.Update
	errors  []error
}

func (s *scriptedSource) GetUpdates(ctx context.Context, params api.GetUpdatesParams) ([]telegram.Update, error) {
	s.mu.Lock()
	index := len(s.calls)
	s.calls = append(s.calls, params)
	var updates []telegram.Update
	var err error
	if index < len(s.results) {
		updates = s.results[index]
	}
	if index < len(s.errors) {
		err = s.errors[index]
	}
	s.mu.Unlock()
	if index >= len(s.results) && index >= len(s.errors) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return updates, err
}

func TestPollAdvancesOffsetAndDrains(t *testing.T) {
	t.Parallel()

	source := &scriptedSource{results: [][]telegram.Update{{
		{UpdateID: 10}, {UpdateID: 12},
	}}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var got []int64
	waited := false
	err := Poll(ctx, source, func(_ context.Context, update *telegram.Update, wait bool) bool {
		if !wait {
			t.Fatal("poll dispatch must apply backpressure")
		}
		got = append(got, update.UpdateID)
		if len(got) == 2 {
			cancel()
		}
		return true
	}, func() { waited = true }, PollOptions{Offset: 5, Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != 10 || got[1] != 12 {
		t.Fatalf("updates = %v", got)
	}
	if !waited {
		t.Fatal("drain callback was not called")
	}
	source.mu.Lock()
	defer source.mu.Unlock()
	if len(source.calls) == 0 || source.calls[0].Offset != 5 {
		t.Fatalf("first poll params = %#v", source.calls)
	}
}

func TestPollRetriesTransientFailure(t *testing.T) {
	t.Parallel()

	source := &scriptedSource{
		results: [][]telegram.Update{nil, {{UpdateID: 1}}},
		errors:  []error{errors.New("temporary"), nil},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	start := time.Now()
	err := Poll(ctx, source, func(_ context.Context, update *telegram.Update, _ bool) bool {
		if update.UpdateID == 1 {
			cancel()
		}
		return true
	}, nil, PollOptions{MinBackoff: time.Millisecond, MaxBackoff: 2 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(start) < time.Millisecond {
		t.Fatal("transient failure was not backed off")
	}
	source.mu.Lock()
	calls := len(source.calls)
	source.mu.Unlock()
	if calls < 2 {
		t.Fatalf("calls = %d", calls)
	}
}

func TestPollSkipsStaleAndDuplicateUpdates(t *testing.T) {
	t.Parallel()

	source := &scriptedSource{results: [][]telegram.Update{{
		{UpdateID: 10}, {UpdateID: 10}, {UpdateID: 9}, {UpdateID: 12},
	}}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var got []int64
	err := Poll(ctx, source, func(_ context.Context, update *telegram.Update, _ bool) bool {
		got = append(got, update.UpdateID)
		if len(got) == 2 {
			cancel()
		}
		return true
	}, nil, PollOptions{Offset: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != 10 || got[1] != 12 {
		t.Fatalf("updates = %v", got)
	}
}

func TestPollValidatesDependencies(t *testing.T) {
	t.Parallel()

	dispatch := func(context.Context, *telegram.Update, bool) bool { return true }
	if err := Poll(context.Background(), nil, dispatch, nil, PollOptions{}); !errors.Is(err, ErrUpdateSourceRequired) {
		t.Fatalf("source error = %v", err)
	}
	if err := Poll(context.Background(), &scriptedSource{}, nil, nil, PollOptions{}); !errors.Is(err, ErrDispatchRequired) {
		t.Fatalf("dispatch error = %v", err)
	}
}

func TestPollBackoffSaturates(t *testing.T) {
	t.Parallel()

	const maximum = time.Duration(1<<63 - 1)
	if got := nextBackoff(maximum/2+1, maximum); got != maximum {
		t.Fatalf("next backoff = %v", got)
	}
	if got := retryAfterDuration(int(^uint(0) >> 1)); got <= 0 || got > maximum {
		t.Fatalf("retry-after duration = %v", got)
	}
	options := (PollOptions{MinBackoff: time.Minute}).normalized()
	if options.MaxBackoff != time.Minute {
		t.Fatalf("normalized max backoff = %v", options.MaxBackoff)
	}
}
