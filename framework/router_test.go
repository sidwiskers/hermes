package framework

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	telegram "github.com/sidwiskers/hermes/types"
)

func messageUpdate(text string) *telegram.Update {
	return &telegram.Update{Message: &telegram.Message{
		MessageID: 1,
		Chat:      telegram.Chat{ID: 7, Type: "private"},
		Text:      text,
	}}
}

func TestRouterDirectCommandAndMiddleware(t *testing.T) {
	t.Parallel()

	router := NewRouter()
	var sequence []string
	router.Use(func(next Handler) Handler {
		return func(c *Context) error {
			sequence = append(sequence, "before")
			err := next(c)
			sequence = append(sequence, "after")
			return err
		}
	})
	router.Command("start", func(c *Context) error {
		sequence = append(sequence, c.Args())
		return nil
	})

	ctx := NewContext(context.Background(), nil, messageUpdate("/start hello"), "")
	if err := router.Handle(ctx); err != nil {
		t.Fatal(err)
	}
	want := []string{"before", "hello", "after"}
	if fmt.Sprint(sequence) != fmt.Sprint(want) {
		t.Fatalf("sequence = %v, want %v", sequence, want)
	}
}

func TestRouterSnapshotConcurrentReadsAndWrites(t *testing.T) {
	router := NewRouter()
	router.Command("start", func(*Context) error { return nil })
	ctx := NewContext(context.Background(), nil, messageUpdate("/start"), "")

	var failed atomic.Bool
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := 0; index < 500; index++ {
				if err := router.Handle(ctx); err != nil {
					failed.Store(true)
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for index := 0; index < 100; index++ {
			router.CallbackPrefix(fmt.Sprintf("item:%d:", index), func(*Context) error { return nil })
		}
	}()
	wg.Wait()
	if failed.Load() {
		t.Fatal("concurrent route read failed")
	}
}

func BenchmarkRouterExactCommand(b *testing.B) {
	router := NewRouter()
	router.Command("start", func(*Context) error { return nil })
	ctx := NewContext(context.Background(), nil, messageUpdate("/start"), "")
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if err := router.Handle(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRouterThousandFilteredRoutes(b *testing.B) {
	router := NewRouter()
	for index := 0; index < 999; index++ {
		value := fmt.Sprintf("no-match-%d", index)
		router.On(TextEquals(value), func(*Context) error { return nil })
	}
	router.On(TextEquals("target"), func(*Context) error { return nil })
	ctx := NewContext(context.Background(), nil, messageUpdate("target"), "")
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if err := router.Handle(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRouterMiddlewareDepth10(b *testing.B) {
	router := NewRouter()
	for index := 0; index < 10; index++ {
		router.Use(func(next Handler) Handler {
			return func(c *Context) error { return next(c) }
		})
	}
	router.Command("start", func(*Context) error { return nil })
	ctx := NewContext(context.Background(), nil, messageUpdate("/start"), "")
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if err := router.Handle(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func TestStartupRegistrationsCompileOnceAndPostStartChangesAreVisible(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	for index := 0; index < 1000; index++ {
		router.Command(fmt.Sprintf("command%d", index), func(*Context) error { return nil })
	}
	called := false
	router.Command("target", func(*Context) error { called = true; return nil })
	ctx := NewContext(context.Background(), nil, messageUpdate("/target"), "")
	if err := router.Handle(ctx); err != nil || !called {
		t.Fatalf("startup route: called=%v err=%v", called, err)
	}

	called = false
	router.Command("later", func(*Context) error { called = true; return nil })
	if err := router.Handle(NewContext(context.Background(), nil, messageUpdate("/later"), "")); err != nil || !called {
		t.Fatalf("post-start route: called=%v err=%v", called, err)
	}
}

func BenchmarkRouterStartupRegistration1000(b *testing.B) {
	for run := 0; run < b.N; run++ {
		router := NewRouter()
		for index := 0; index < 1000; index++ {
			router.Command(fmt.Sprintf("command%d", index), func(*Context) error { return nil })
		}
		if err := router.Handle(NewContext(context.Background(), nil, messageUpdate("/command999"), "")); err != nil {
			b.Fatal(err)
		}
	}
}
