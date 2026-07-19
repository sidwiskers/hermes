package framework

import (
	"context"
	"errors"
	"testing"

	telegram "github.com/sidwiskers/hermes/types"
)

func TestContextDirectWebhookResponse(t *testing.T) {
	t.Parallel()
	ctx := NewContext(context.Background(), nil, &telegram.Update{}, "")
	params := map[string]any{"chat_id": int64(1), "text": "hello"}
	if err := ctx.RespondWebhook("sendMessage", params); err != nil {
		t.Fatal(err)
	}
	response, ok := ctx.DirectWebhookResponse()
	if !ok || response.Method != "sendMessage" || response.Params == nil {
		t.Fatalf("response=%+v ok=%v", response, ok)
	}
	if err := ctx.RespondWebhook("deleteMessage", nil); !errors.Is(err, ErrWebhookResponseSet) {
		t.Fatalf("duplicate error=%v", err)
	}
	for _, method := range []string{"", "../sendMessage", "send message"} {
		fresh := NewContext(context.Background(), nil, &telegram.Update{}, "")
		if err := fresh.RespondWebhook(method, nil); !errors.Is(err, ErrWebhookMethod) {
			t.Fatalf("method %q error=%v", method, err)
		}
	}
}

func TestContextCopiesShareWebhookResponseButCloneIsIndependent(t *testing.T) {
	original := NewContext(context.Background(), nil, nil, "")
	derived := *original
	if err := derived.RespondWebhook("sendMessage", struct {
		Text string `json:"text"`
	}{Text: "shared"}); err != nil {
		t.Fatal(err)
	}
	if response, ok := original.DirectWebhookResponse(); !ok || response.Method != "sendMessage" {
		t.Fatalf("shared response = %#v, %v", response, ok)
	}

	clone := original.Clone()
	if _, ok := clone.DirectWebhookResponse(); !ok {
		t.Fatal("clone did not preserve the response present when cloned")
	}
	if err := clone.RespondWebhook("deleteMessage", nil); !errors.Is(err, ErrWebhookResponseSet) {
		t.Fatalf("duplicate clone response error = %v", err)
	}

	base := NewContext(context.Background(), nil, nil, "")
	empty := base.Clone()
	if err := empty.RespondWebhook("deleteMessage", nil); err != nil {
		t.Fatal(err)
	}
	if _, ok := base.DirectWebhookResponse(); ok {
		t.Fatal("independent context unexpectedly shared response state")
	}
}

func TestContextCommandMentionRules(t *testing.T) {
	t.Parallel()

	update := &telegram.Update{Message: &telegram.Message{
		Chat: telegram.Chat{ID: 1, Type: "private"},
		Text: "/Deploy@My_Bot now",
	}}
	ctx := NewContext(context.Background(), nil, update, "my_bot")
	if ctx.Command() != "deploy" || ctx.Args() != "now" {
		t.Fatalf("command=%q args=%q", ctx.Command(), ctx.Args())
	}

	ignored := NewContext(context.Background(), nil, update, "other_bot")
	if ignored.Command() != "" {
		t.Fatalf("routed foreign mention %q", ignored.Command())
	}
}

func FuzzCommandParsing(f *testing.F) {
	f.Add("/start hello", "bot")
	f.Add("/start@bot hello", "bot")
	f.Add("text", "bot")
	f.Fuzz(func(t *testing.T, text, username string) {
		message := &telegram.Message{Text: text}
		command, _ := parseCommand(message, username)
		for _, char := range command {
			if char >= 'A' && char <= 'Z' {
				t.Fatalf("command not normalized: %q", command)
			}
		}
	})
}
