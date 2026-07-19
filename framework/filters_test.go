package framework

import (
	"context"
	"testing"

	telegram "github.com/sidwiskers/hermes/types"
)

func filterContext() *Context {
	return NewContext(context.Background(), nil, &telegram.Update{Message: &telegram.Message{
		MessageID: 3,
		From:      &telegram.User{ID: 11, FirstName: "Ada"},
		Chat:      telegram.Chat{ID: 22, Type: "private"},
		Text:      "hello world",
	}}, "")
}

func TestComposableFilters(t *testing.T) {
	t.Parallel()
	ctx := filterContext()

	checks := []struct {
		name string
		got  bool
		want bool
	}{
		{"all", All(TextMessage, PrivateChat)(ctx), true},
		{"any", Any(GroupChat, PrivateChat)(ctx), true},
		{"not", Not(GroupChat)(ctx), true},
		{"type", UpdateIs(telegram.UpdateMessage)(ctx), true},
		{"sender", FromUsers(11)(ctx), true},
		{"chat", InChats(22)(ctx), true},
		{"equals", TextEquals("hello world")(ctx), true},
		{"prefix", TextPrefix("hello")(ctx), true},
		{"group false", GroupChat(ctx), false},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %v, want %v", check.name, check.got, check.want)
		}
	}
}

func TestCallbackFilters(t *testing.T) {
	t.Parallel()
	ctx := NewContext(context.Background(), nil, &telegram.Update{CallbackQuery: &telegram.CallbackQuery{
		ID: "cb", Data: "user:42", From: telegram.User{ID: 11},
		Message: telegram.AccessibleMessage(&telegram.Message{Chat: telegram.Chat{ID: 22, Type: "private"}}),
	}}, "")
	if !CallbackUpdate(ctx) || !CallbackDataPrefix("user:")(ctx) || !MessageUpdate(ctx) {
		t.Fatalf("callback filters failed")
	}
}
