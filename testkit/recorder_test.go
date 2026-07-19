package testkit

import (
	"context"
	"testing"

	"github.com/sidwiskers/hermes"
)

func TestRecorder(t *testing.T) {
	bot, recorder := New()
	recorder.Respond(hermes.Message{MessageID: 9, Chat: hermes.Chat{ID: 7, Type: "private"}})

	message, err := bot.SendMessage(context.Background(), hermes.SendMessageParams{ChatID: 7, Text: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if message.MessageID != 9 {
		t.Fatalf("message = %#v", message)
	}
	request, ok := recorder.Last()
	if !ok || request.Method != "sendMessage" || request.JSON["text"] != "hello" {
		t.Fatalf("request = %#v", request)
	}
}

func TestRecorderStandaloneClient(t *testing.T) {
	client, recorder := NewClient()
	recorder.Respond(hermes.Message{MessageID: 11, Chat: hermes.Chat{ID: 8, Type: "private"}})

	message, err := client.SendMessage(context.Background(), hermes.SendMessageParams{ChatID: 8, Text: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if message.MessageID != 11 {
		t.Fatalf("message = %#v", message)
	}
}
