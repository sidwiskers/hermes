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

func TestCreateForumTopic(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/botTOKEN/createForumTopic" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		var params CreateForumTopicParams
		if err := json.NewDecoder(request.Body).Decode(&params); err != nil {
			t.Fatal(err)
		}
		if params.Name != "Release" || params.IconColor != ForumIconBlue {
			t.Fatalf("params = %#v", params)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"message_thread_id":42,"name":"Release","icon_color":7322096}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	topic, err := client.CreateForumTopic(context.Background(), CreateForumTopicParams{
		ChatID: 7, Name: "Release", IconColor: ForumIconBlue,
	})
	if err != nil {
		t.Fatal(err)
	}
	if topic.MessageThreadID != 42 {
		t.Fatalf("topic = %#v", topic)
	}
}

func TestEditForumTopicCanRemoveIcon(t *testing.T) {
	t.Parallel()

	empty := ""
	data, err := json.Marshal(EditForumTopicParams{
		ChatID: 7, MessageThreadID: 42, IconCustomEmojiID: &empty,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"icon_custom_emoji_id":""`) {
		t.Fatalf("params = %s", data)
	}
}

func TestForumValidation(t *testing.T) {
	t.Parallel()

	client := New("TOKEN")
	_, err := client.CreateForumTopic(context.Background(), CreateForumTopicParams{
		ChatID: 1, Name: "x", IconColor: 1,
	})
	if err == nil {
		t.Fatal("expected invalid icon color")
	}
	if err := client.CloseForumTopic(context.Background(), ForumTopicTargetParams{ChatID: 1}); err == nil {
		t.Fatal("expected missing message_thread_id")
	}
}
