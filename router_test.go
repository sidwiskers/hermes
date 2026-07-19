package hermes

import (
	"context"
	"reflect"
	"testing"
)

func TestCommandRoutingAndMiddlewareOrder(t *testing.T) {
	t.Parallel()

	bot := New("TOKEN", WithBotUsername("sample_bot"))
	order := make([]string, 0, 3)

	bot.Use(func(next Handler) Handler {
		return func(c *Context) error {
			order = append(order, "before")
			err := next(c)
			order = append(order, "after")
			return err
		}
	})
	bot.Command("start", func(c *Context) error {
		order = append(order, c.Args())
		return nil
	})

	update := &Update{Message: &Message{
		MessageID: 1,
		Chat:      Chat{ID: 9, Type: "private"},
		Text:      "/start@sample_bot one two",
	}}
	if err := bot.Handle(context.Background(), update); err != nil {
		t.Fatal(err)
	}

	want := []string{"before", "one two", "after"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("order = %#v, want %#v", order, want)
	}
}

func TestCommandForAnotherBotIsIgnored(t *testing.T) {
	t.Parallel()

	bot := New("TOKEN", WithBotUsername("ours"))
	called := false
	bot.Command("start", func(*Context) error {
		called = true
		return nil
	})

	_ = bot.Handle(context.Background(), &Update{Message: &Message{
		Chat: Chat{ID: 1},
		Text: "/start@theirs",
	}})
	if called {
		t.Fatal("routed a command addressed to another bot")
	}
}

func TestLongestCallbackPrefixWins(t *testing.T) {
	t.Parallel()

	bot := New("TOKEN")
	got := ""
	bot.CallbackPrefix("user:", func(*Context) error {
		got = "short"
		return nil
	})
	bot.CallbackPrefix("user:ban:", func(*Context) error {
		got = "long"
		return nil
	})

	update := &Update{CallbackQuery: &CallbackQuery{
		ID:      "1",
		Data:    "user:ban:42",
		From:    User{ID: 7},
		Message: AccessibleMessage(&Message{Chat: Chat{ID: 9}}),
	}}
	if err := bot.Handle(context.Background(), update); err != nil {
		t.Fatal(err)
	}
	if got != "long" {
		t.Fatalf("got %q", got)
	}
}

func BenchmarkRouterCommand(b *testing.B) {
	bot := New("TOKEN")
	bot.Command("start", func(*Context) error { return nil })
	update := &Update{Message: &Message{
		Chat: Chat{ID: 1},
		Text: "/start",
	}}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := bot.Handle(ctx, update); err != nil {
			b.Fatal(err)
		}
	}
}

func TestConcurrentRegistrationAndDispatch(t *testing.T) {
	bot := New("TOKEN")
	bot.Command("start", func(*Context) error { return nil })
	update := &Update{Message: &Message{Chat: Chat{ID: 1}, Text: "/start"}}

	const loops = 500
	done := make(chan struct{})
	go func() {
		defer close(done)
		for index := 0; index < loops; index++ {
			bot.CallbackPrefix("item:", func(*Context) error { return nil })
		}
	}()

	for index := 0; index < loops; index++ {
		if err := bot.Handle(context.Background(), update); err != nil {
			t.Fatal(err)
		}
	}
	<-done
}
