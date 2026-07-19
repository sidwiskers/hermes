package framework

import (
	"context"
	"testing"

	telegram "github.com/sidwiskers/hermes/types"
)

func TestContextPoolClearsBorrowedState(t *testing.T) {
	t.Parallel()

	pool := NewContextPool()
	first := pool.Acquire(context.Background(), nil, messageUpdate("/start old"), "")
	if first.Command() != "start" || first.Args() != "old" {
		t.Fatalf("first context = command %q args %q", first.Command(), first.Args())
	}
	pool.Release(first)

	second := pool.Acquire(context.Background(), nil, &telegram.Update{UpdateID: 2}, "")
	defer pool.Release(second)
	if second.Command() != "" || second.Args() != "" || second.Message != nil || second.Update.UpdateID != 2 {
		t.Fatalf("stale pooled state: %#v", second)
	}
}

func TestContextCloneSurvivesPoolRelease(t *testing.T) {
	t.Parallel()
	pool := NewContextPool()
	borrowed := pool.Acquire(context.Background(), nil, messageUpdate("/start keep"), "")
	cloned := borrowed.Clone()
	pool.Release(borrowed)
	if cloned.Command() != "start" || cloned.Args() != "keep" || cloned.Message == nil {
		t.Fatalf("cloned context lost state: %#v", cloned)
	}
}

func TestContextPoolClearsWebhookResponse(t *testing.T) {
	t.Parallel()
	pool := NewContextPool()
	first := pool.Acquire(context.Background(), nil, &telegram.Update{}, "")
	if err := first.RespondWebhook("sendMessage", map[string]any{"text": "secret"}); err != nil {
		t.Fatal(err)
	}
	pool.Release(first)
	second := pool.Acquire(context.Background(), nil, &telegram.Update{}, "")
	defer pool.Release(second)
	if response, ok := second.DirectWebhookResponse(); ok {
		t.Fatalf("pooled response leaked: %+v", response)
	}
}
