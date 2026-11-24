package framework

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sidwiskers/hermes/api"
	telegram "github.com/sidwiskers/hermes/types"
)

func TestInlineCallbackModeKeepsEditHelpersUsableWithoutMessage(t *testing.T) {
	t.Parallel()

	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		methods = append(methods, request.URL.Path)
		var body map[string]json.RawMessage
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		var inlineID string
		if err := json.Unmarshal(body["inline_message_id"], &inlineID); err != nil {
			t.Fatal(err)
		}
		if inlineID != "inline-123" {
			t.Fatalf("%s inline_message_id = %q", request.URL.Path, inlineID)
		}
		for _, forbidden := range []string{"chat_id", "message_id"} {
			if _, exists := body[forbidden]; exists {
				t.Fatalf("%s unexpectedly included %s: %#v", request.URL.Path, forbidden, body)
			}
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
	}))
	defer server.Close()

	client := api.New("TOKEN", api.WithBaseURL(server.URL), api.WithHTTPClient(server.Client()))
	update := &telegram.Update{
		CallbackQuery: &telegram.CallbackQuery{
			ID:              "callback-1",
			From:            telegram.User{ID: 42, FirstName: "User"},
			InlineMessageID: "inline-123",
		},
	}
	ctx := NewContext(context.Background(), client, update, "")

	if ctx.Message != nil {
		t.Fatalf("Message = %#v, want nil for inline callback mode", ctx.Message)
	}
	if _, ok := ctx.ChatID(); ok {
		t.Fatal("ChatID unexpectedly available without CallbackQuery.Message")
	}
	if _, ok := ctx.MessageID(); ok {
		t.Fatal("MessageID unexpectedly available without CallbackQuery.Message")
	}
	if got, ok := ctx.InlineMessageID(); !ok || got != "inline-123" {
		t.Fatalf("InlineMessageID = %q, %v", got, ok)
	}
	if err := ctx.Edit("updated"); err != nil {
		t.Fatal(err)
	}
	if err := ctx.EditCaption("caption"); err != nil {
		t.Fatal(err)
	}
	if err := ctx.EditKeyboard(nil); err != nil {
		t.Fatal(err)
	}
	if err := ctx.Delete(); err == nil {
		t.Fatal("Delete should require an ordinary Message")
	}
	if err := ctx.React(ReactionType{}); err == nil {
		t.Fatal("React should require an ordinary Message")
	}
	if len(methods) != 3 || !strings.HasSuffix(methods[0], "/editMessageText") ||
		!strings.HasSuffix(methods[1], "/editMessageCaption") ||
		!strings.HasSuffix(methods[2], "/editMessageReplyMarkup") {
		t.Fatalf("edit methods = %#v", methods)
	}
}
