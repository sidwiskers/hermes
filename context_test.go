package hermes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	telegram "github.com/sidwiskers/hermes/types"
)

func TestContextEphemeralFromCallback(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var params SendMessageParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatal(err)
		}
		if params.ChatID.(float64) != 100 || params.ReceiverUserID != 7 || params.CallbackQueryID != "cb" {
			t.Fatalf("unexpected params: %#v", params)
		}
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":0,"chat":{"id":100,"type":"supergroup"},"receiver_user":{"id":7,"is_bot":false,"first_name":"A"},"ephemeral_message_id":9}}`))
	}))
	defer server.Close()

	bot := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	update := &Update{CallbackQuery: &CallbackQuery{
		ID:   "cb",
		From: User{ID: 7, FirstName: "A"},
		Message: AccessibleMessage(&Message{
			Chat: Chat{ID: 100, Type: "supergroup"},
		}),
	}}
	c := newContext(context.Background(), bot, update)
	message, err := c.EphemeralMessage("private")
	if err != nil {
		t.Fatal(err)
	}
	if message.EphemeralMessageID != 9 {
		t.Fatalf("unexpected message: %#v", message)
	}
}

func TestContextEphemeralRepliesToEphemeralCommand(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var params SendMessageParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatal(err)
		}
		if params.CallbackQueryID != "" {
			t.Fatalf("unexpected callback id: %q", params.CallbackQueryID)
		}
		if params.ReplyParameters == nil || params.ReplyParameters.EphemeralMessageID != 22 {
			t.Fatalf("missing ephemeral reply: %#v", params)
		}
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":0,"chat":{"id":100,"type":"supergroup"},"receiver_user":{"id":7,"is_bot":false,"first_name":"A"},"ephemeral_message_id":23}}`))
	}))
	defer server.Close()

	bot := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	update := &Update{Message: &Message{
		MessageID:          0,
		EphemeralMessageID: 22,
		From:               &User{ID: 7, FirstName: "A"},
		Chat:               Chat{ID: 100, Type: "supergroup"},
		Text:               "/profile",
	}}
	c := newContext(context.Background(), bot, update)
	if err := c.Ephemeral("private"); err != nil {
		t.Fatal(err)
	}
}

func TestReplyUsesNormalMessageID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var params SendMessageParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatal(err)
		}
		if params.ReplyParameters == nil || params.ReplyParameters.MessageID != 44 {
			t.Fatalf("missing normal reply: %#v", params)
		}
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":45,"chat":{"id":100,"type":"supergroup"}}}`))
	}))
	defer server.Close()

	bot := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	c := newContext(context.Background(), bot, &Update{Message: &Message{
		MessageID: 44,
		Chat:      Chat{ID: 100, Type: "supergroup"},
	}})
	if err := c.Reply("hello"); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateRawJSONIsOptIn(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"update_id":9,"future_update":{"value":1}}`)
	var fast Update
	if err := json.Unmarshal(payload, &fast); err != nil {
		t.Fatal(err)
	}
	if fast.UpdateID != 9 || len(fast.Raw) != 0 {
		t.Fatalf("fast decode unexpectedly retained raw JSON: %#v", fast)
	}

	preserved, err := telegram.DecodeUpdate(payload, true)
	if err != nil {
		t.Fatal(err)
	}
	if preserved.UpdateID != 9 || !bytes.Equal(preserved.Raw, payload) {
		t.Fatalf("raw update not retained: %#v", preserved)
	}
}
