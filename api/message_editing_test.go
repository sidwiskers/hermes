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

func TestEditMessageMediaInjectsVariantType(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		media := body["media"].(map[string]any)
		if media["type"] != "photo" || media["media"] != "photo-id" {
			t.Fatalf("media = %#v", media)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"message_id":2,"chat":{"id":1,"type":"private"}}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	message, err := client.EditMessageMedia(context.Background(), EditMessageMediaParams{
		ChatID: 1, MessageID: 2, Media: InputMediaPhoto{Media: "photo-id"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if message == nil || message.MessageID != 2 {
		t.Fatalf("message = %#v", message)
	}
}

func TestEditMessageMediaStreamsNestedAttachments(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		media := request.FormValue("media")
		if !strings.Contains(media, `"type":"live_photo"`) || !strings.Contains(media, "attach://still") {
			t.Fatalf("media = %s", media)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"message_id":2,"chat":{"id":1,"type":"private"}}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	_, err := client.EditMessageMediaUpload(context.Background(), EditMessageMediaParams{
		ChatID: 1, MessageID: 2,
		Media: InputMediaLivePhoto{Media: "video-id", Photo: Attachment("still")},
	}, NewUpload("still", "still.jpg", strings.NewReader("image")))
	if err != nil {
		t.Fatal(err)
	}
}

func TestMessageEditTargetValidation(t *testing.T) {
	t.Parallel()

	client := New("TOKEN")
	_, err := client.StopMessageLiveLocation(context.Background(), StopMessageLiveLocationParams{
		ChatID: 1, MessageID: 2, InlineMessageID: "inline",
	})
	if err == nil {
		t.Fatal("expected conflicting target error")
	}
}
