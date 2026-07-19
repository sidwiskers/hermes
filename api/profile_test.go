package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetBusinessConnectionDecodesRights(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/botTOKEN/getBusinessConnection" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"id":"business-1","user":{"id":9,"is_bot":false,"first_name":"Ada"},"user_chat_id":9,"date":1,"rights":{"can_reply":true,"can_manage_stories":true},"is_enabled":true}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	connection, err := client.GetBusinessConnection(context.Background(), GetBusinessConnectionParams{BusinessConnectionID: "business-1"})
	if err != nil {
		t.Fatal(err)
	}
	if connection.Rights == nil || !connection.Rights.CanReply || !connection.Rights.CanManageStories {
		t.Fatalf("connection = %#v", connection)
	}
}

func TestSetChatMenuButtonContract(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		button := body["menu_button"].(map[string]any)
		if button["type"] != MenuButtonTypeWebApp || button["text"] != "Open" {
			t.Fatalf("menu button = %#v", button)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	button := WebAppMenuButton("Open", "https://example.com/app")
	if err := client.SetChatMenuButton(context.Background(), SetChatMenuButtonParams{ChatID: 7, MenuButton: &button}); err != nil {
		t.Fatal(err)
	}
}

func TestProfileValidation(t *testing.T) {
	t.Parallel()

	client := New("TOKEN")
	if _, err := client.GetUserProfilePhotos(context.Background(), GetUserProfilePhotosParams{UserID: 1, Limit: 101}); err == nil {
		t.Fatal("expected invalid limit")
	}
	if err := client.VerifyUser(context.Background(), VerifyUserParams{UserID: 1, CustomDescription: string(make([]rune, 71))}); err == nil {
		t.Fatal("expected long custom description")
	}
	if err := client.SetChatMenuButton(context.Background(), SetChatMenuButtonParams{MenuButton: &MenuButton{Type: MenuButtonTypeWebApp}}); err == nil {
		t.Fatal("expected incomplete web app menu button")
	}
}
