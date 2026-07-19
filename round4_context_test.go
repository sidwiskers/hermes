package hermes

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContextRound4Actions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		update *Update
		call   func(*Context) error
		check  func(*testing.T, map[string]any)
	}{
		{
			name:   "reaction",
			method: "setMessageReaction",
			update: &Update{Message: &Message{MessageID: 44, Chat: Chat{ID: -1001, Type: "supergroup"}}},
			call: func(c *Context) error {
				return c.React(EmojiReaction("🔥"), true)
			},
			check: func(t *testing.T, body map[string]any) {
				reactions, ok := body["reaction"].([]any)
				if !ok || len(reactions) != 1 || reactions[0].(map[string]any)["emoji"] != "🔥" {
					t.Fatalf("reaction = %#v", body["reaction"])
				}
				if body["is_big"] != true {
					t.Fatalf("is_big = %#v", body["is_big"])
				}
			},
		},
		{
			name:   "approve join request",
			method: "approveChatJoinRequest",
			update: &Update{ChatJoinRequest: &ChatJoinRequest{Chat: Chat{ID: -1002, Type: "supergroup"}, From: User{ID: 77}}},
			call:   func(c *Context) error { return c.ApproveJoinRequest() },
			check: func(t *testing.T, body map[string]any) {
				if body["chat_id"] != float64(-1002) || body["user_id"] != float64(77) {
					t.Fatalf("join request = %#v", body)
				}
			},
		},
		{
			name:   "poll",
			method: "sendPoll",
			update: &Update{Message: &Message{MessageID: 1, Chat: Chat{ID: 99, Type: "private"}}},
			call:   func(c *Context) error { return c.Poll("Pick one", "A", "B") },
			check: func(t *testing.T, body map[string]any) {
				if body["question"] != "Pick one" {
					t.Fatalf("question = %#v", body["question"])
				}
				options, ok := body["options"].([]any)
				if !ok || len(options) != 2 {
					t.Fatalf("options = %#v", body["options"])
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasSuffix(r.URL.Path, "/"+test.method) {
					t.Fatalf("path = %s", r.URL.Path)
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				test.check(t, body)
				result := "true"
				if test.method == "sendPoll" {
					result = `{"message_id":2,"chat":{"id":99,"type":"private"},"poll":{"id":"p","question":"Pick one","options":[],"total_voter_count":0,"is_closed":false,"is_anonymous":true,"type":"regular","allows_multiple_answers":false}}`
				}
				_, _ = io.WriteString(w, `{"ok":true,"result":`+result+`}`)
			}))
			defer server.Close()

			bot := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
			if err := test.call(newContext(context.Background(), bot, test.update)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestContextReactRejectsNonMessageUpdates(t *testing.T) {
	t.Parallel()

	bot := New("TOKEN")
	ctx := newContext(context.Background(), bot, &Update{CallbackQuery: &CallbackQuery{ID: "x", From: User{ID: 1}}})
	if err := ctx.React(EmojiReaction("👍")); err == nil {
		t.Fatal("expected missing-message error")
	}
}
