package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBotAPI103ChangedMethodParameters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		call   func(*Client) error
		check  func(*testing.T, map[string]any)
	}{
		{
			name: "message draft stop controls", method: "sendMessageDraft",
			call: func(client *Client) error {
				return client.SendMessageDraft(context.Background(), SendMessageDraftParams{
					ChatID: 1, DraftID: 2, Text: "partial", CanStop: true, KeepOnStop: true,
				})
			},
			check: func(t *testing.T, body map[string]any) {
				if body["can_stop"] != true || body["keep_on_stop"] != true {
					t.Fatalf("body = %#v", body)
				}
			},
		},
		{
			name: "rich draft stop controls", method: "sendRichMessageDraft",
			call: func(client *Client) error {
				return client.SendRichMessageDraft(context.Background(), SendRichMessageDraftParams{
					ChatID: 1, DraftID: 2, RichMessage: InputRichMessage{HTML: "partial"}, CanStop: true, KeepOnStop: true,
				})
			},
			check: func(t *testing.T, body map[string]any) {
				if body["can_stop"] != true || body["keep_on_stop"] != true {
					t.Fatalf("body = %#v", body)
				}
			},
		},
		{
			name: "ephemeral rich edit", method: "editEphemeralMessageText",
			call: func(client *Client) error {
				return client.EditEphemeralText(context.Background(), EditEphemeralMessageTextParams{
					ChatID: 1, ReceiverUserID: 2, EphemeralMessageID: 3,
					RichMessage: &InputRichMessage{HTML: "<b>updated</b>"},
				})
			},
			check: func(t *testing.T, body map[string]any) {
				if _, ok := body["rich_message"].(map[string]any); !ok {
					t.Fatalf("body = %#v", body)
				}
				if _, exists := body["text"]; exists {
					t.Fatalf("empty text leaked to body: %#v", body)
				}
			},
		},
		{
			name: "ephemeral caption placement", method: "editEphemeralMessageCaption",
			call: func(client *Client) error {
				return client.EditEphemeralCaption(context.Background(), EditEphemeralMessageCaptionParams{
					ChatID: 1, ReceiverUserID: 2, EphemeralMessageID: 3, ShowCaptionAboveMedia: true,
				})
			},
			check: func(t *testing.T, body map[string]any) {
				if body["show_caption_above_media"] != true {
					t.Fatalf("body = %#v", body)
				}
			},
		},
		{
			name: "welcome permission", method: "promoteChatMember",
			call: func(client *Client) error {
				return client.PromoteChatMember(context.Background(), PromoteChatMemberParams{
					ChatID: 1, UserID: 2, CanSendWelcomeMessages: true,
				})
			},
			check: func(t *testing.T, body map[string]any) {
				if body["can_send_welcome_messages"] != true {
					t.Fatalf("body = %#v", body)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if !strings.HasSuffix(request.URL.Path, "/"+test.method) {
					t.Fatalf("path = %s", request.URL.Path)
				}
				var body map[string]any
				if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				test.check(t, body)
				_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
			}))
			defer server.Close()

			client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
			if err := test.call(client); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEditEphemeralMediaUpload(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		media := request.FormValue("media")
		if !strings.Contains(media, `"type":"document"`) || !strings.Contains(media, `"media":"attach://document"`) {
			t.Fatalf("media = %s", media)
		}
		file, _, err := request.FormFile("document")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "document-data" {
			t.Fatalf("upload = %q", data)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	err := client.EditEphemeralMediaUpload(context.Background(), EditEphemeralMessageMediaParams{
		ChatID: 1, ReceiverUserID: 2, EphemeralMessageID: 3,
		TypedMedia: InputMediaDocument{Media: Attachment("document")},
	}, NewUpload("document", "document.txt", strings.NewReader("document-data")))
	if err != nil {
		t.Fatal(err)
	}
}
