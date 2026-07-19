package hermes

import (
	"context"
	"reflect"
	"testing"
)

func TestFilteredRoutesAndGroups(t *testing.T) {
	t.Parallel()

	bot := New("TOKEN")
	order := []string{}
	private := bot.Group(PrivateChat).Use(func(next Handler) Handler {
		return func(c *Context) error {
			order = append(order, "group-before")
			err := next(c)
			order = append(order, "group-after")
			return err
		}
	})
	private.Command("settings", func(*Context) error {
		order = append(order, "handler")
		return nil
	})

	groupUpdate := &Update{Message: &Message{Chat: Chat{ID: -100, Type: "supergroup"}, Text: "/settings"}}
	if err := bot.Handle(context.Background(), groupUpdate); err != nil {
		t.Fatal(err)
	}
	if len(order) != 0 {
		t.Fatalf("private route matched group: %#v", order)
	}

	privateUpdate := &Update{Message: &Message{Chat: Chat{ID: 7, Type: "private"}, Text: "/settings"}}
	if err := bot.Handle(context.Background(), privateUpdate); err != nil {
		t.Fatal(err)
	}
	want := []string{"group-before", "handler", "group-after"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("order = %#v, want %#v", order, want)
	}
}

func TestOrderedFilterRoutes(t *testing.T) {
	t.Parallel()

	bot := New("TOKEN")
	got := ""
	bot.On(All(TextMessage, TextPrefix("hello")), func(*Context) error {
		got = "specific"
		return nil
	})
	bot.On(TextMessage, func(*Context) error {
		got = "general"
		return nil
	})

	if err := bot.Handle(context.Background(), &Update{Message: &Message{
		Chat: Chat{ID: 1, Type: "private"}, Text: "hello world",
	}}); err != nil {
		t.Fatal(err)
	}
	if got != "specific" {
		t.Fatalf("got %q", got)
	}
}

func TestUpdateTypeAndSenderCoverage(t *testing.T) {
	t.Parallel()

	update := &Update{Subscription: &BotSubscriptionUpdated{
		User: User{ID: 55, FirstName: "Ada"}, State: "active",
	}}
	if update.Type() != UpdateSubscription {
		t.Fatalf("type = %q", update.Type())
	}
	if sender := update.Sender(); sender == nil || sender.ID != 55 {
		t.Fatalf("sender = %#v", sender)
	}
}

func BenchmarkRouterFilteredMessage(b *testing.B) {
	bot := New("TOKEN")
	bot.On(All(TextMessage, PrivateChat), func(*Context) error { return nil })
	update := &Update{Message: &Message{Chat: Chat{ID: 1, Type: "private"}, Text: "hello"}}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := bot.Handle(ctx, update); err != nil {
			b.Fatal(err)
		}
	}
}

func TestCallbackDoesNotRouteCommandFromKeyboardMessage(t *testing.T) {
	t.Parallel()

	bot := New("TOKEN")
	got := ""
	bot.Command("start", func(*Context) error { got = "command"; return nil })
	bot.Callback("button", func(*Context) error { got = "callback"; return nil })

	update := &Update{CallbackQuery: &CallbackQuery{
		ID: "cb", Data: "button", From: User{ID: 1},
		Message: AccessibleMessage(&Message{Chat: Chat{ID: 2, Type: "private"}, Text: "/start"}),
	}}
	if err := bot.Handle(context.Background(), update); err != nil {
		t.Fatal(err)
	}
	if got != "callback" {
		t.Fatalf("got %q", got)
	}
}
