package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestForwardMessagesContract(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/botTOKEN/forwardMessages" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		var params ForwardMessagesParams
		if err := json.NewDecoder(request.Body).Decode(&params); err != nil {
			t.Fatal(err)
		}
		if len(params.MessageIDs) != 3 || params.MessageIDs[2] != 9 || params.DirectMessagesTopicID != 4 {
			t.Fatalf("params = %#v", params)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":[{"message_id":20},{"message_id":21}]}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	messageIDs, err := client.ForwardMessages(context.Background(), ForwardMessagesParams{
		ChatID: 7, FromChatID: 8, MessageIDs: []int{1, 5, 9}, DirectMessagesTopicID: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(messageIDs) != 2 || messageIDs[1].MessageID != 21 {
		t.Fatalf("message IDs = %#v", messageIDs)
	}
}

func TestBulkMessageValidation(t *testing.T) {
	t.Parallel()

	client := New("TOKEN")
	_, err := client.CopyMessages(context.Background(), CopyMessagesParams{
		ChatID: 1, FromChatID: 2, MessageIDs: []int{3, 3},
	})
	if err == nil {
		t.Fatal("expected strictly increasing validation error")
	}
	if err := client.SendMessageDraft(context.Background(), SendMessageDraftParams{ChatID: 1}); err == nil {
		t.Fatal("expected missing draft_id error")
	}
}

func TestLogOutUsesParameterlessMethod(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/botTOKEN/logOut" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	if err := client.LogOut(context.Background()); err != nil {
		t.Fatal(err)
	}
}
