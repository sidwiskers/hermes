package api

import (
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBotAPI104WidgetAndGiftMethods(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body := make(map[string]any)
		rawBody, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		if len(rawBody) != 0 {
			if err := json.Unmarshal(rawBody, &body); err != nil {
				t.Fatal(err)
			}
		}
		switch {
		case strings.HasSuffix(request.URL.Path, "/sendWidget"):
			if body["html"] != "<main>hello</main>" || body["allow_queries"] != true || body["chat_id"] != float64(7) {
				t.Fatalf("sendWidget body = %#v", body)
			}
			_, _ = io.WriteString(writer, `{"ok":true,"result":{"message_id":1,"chat":{"id":7,"type":"private"}}}`)
		case strings.HasSuffix(request.URL.Path, "/sendSharedWidget"):
			if body["widget_id"] != "shared-1" || body["chat_data"] != "session" {
				t.Fatalf("sendSharedWidget body = %#v", body)
			}
			_, _ = io.WriteString(writer, `{"ok":true,"result":{"message_id":2,"chat":{"id":7,"type":"private"}}}`)
		case strings.HasSuffix(request.URL.Path, "/answerWidgetQuery"):
			if body["widget_query_id"] != "query-1" || body["data"] != "answer" {
				t.Fatalf("answerWidgetQuery body = %#v", body)
			}
			_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
		case strings.HasSuffix(request.URL.Path, "/editMessageWidgetData"):
			if body["inline_message_id"] != "inline-1" || body["public_data"] != "" {
				t.Fatalf("editMessageWidgetData body = %#v", body)
			}
			if _, exists := body["chat_data"]; exists {
				t.Fatalf("chat_data should be omitted: %#v", body)
			}
			_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
		case strings.HasSuffix(request.URL.Path, "/editSharedWidgetData"):
			if body["widget_id"] != "shared-1" || body["public_data"] != "new-state" || body["cache_time"] != float64(300) {
				t.Fatalf("editSharedWidgetData body = %#v", body)
			}
			_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
		case strings.HasSuffix(request.URL.Path, "/getUniqueGift"):
			if body["name"] != "gift-name" {
				t.Fatalf("getUniqueGift body = %#v", body)
			}
			_, _ = io.WriteString(writer, `{"ok":true,"result":{}}`)
		case strings.HasSuffix(request.URL.Path, "/getChatShareThemeBackdrops"):
			_, _ = io.WriteString(writer, `{"ok":true,"result":{"backdrops":[]}}`)
		case strings.HasSuffix(request.URL.Path, "/deleteBusinessAccountGift"):
			if body["business_connection_id"] != "business-1" || body["owned_gift_id"] != "owned-1" {
				t.Fatalf("deleteBusinessAccountGift body = %#v", body)
			}
			_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
		case strings.HasSuffix(request.URL.Path, "/verifyUser"):
			if body["custom_description_parse_mode"] != "HTML" || body["custom_description_entities"] == nil {
				t.Fatalf("verifyUser body = %#v", body)
			}
			_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
		case strings.HasSuffix(request.URL.Path, "/verifyChat"):
			if body["custom_description_parse_mode"] != "HTML" || body["custom_description_entities"] == nil {
				t.Fatalf("verifyChat body = %#v", body)
			}
			_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
		default:
			t.Fatalf("unexpected method path %s", request.URL.Path)
		}
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	ctx := context.Background()
	if _, err := client.SendWidget(ctx, SendWidgetParams{ChatID: 7, HTML: "<main>hello</main>", AllowQueries: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.SendSharedWidget(ctx, SendSharedWidgetParams{ChatID: 7, WidgetID: "shared-1", ChatData: "session"}); err != nil {
		t.Fatal(err)
	}
	if err := client.AnswerWidgetQuery(ctx, AnswerWidgetQueryParams{WidgetQueryID: "query-1", Data: "answer"}); err != nil {
		t.Fatal(err)
	}
	empty := ""
	if err := client.EditMessageWidgetData(ctx, EditMessageWidgetDataParams{InlineMessageID: "inline-1", PublicData: &empty}); err != nil {
		t.Fatal(err)
	}
	if err := client.EditSharedWidgetData(ctx, EditSharedWidgetDataParams{WidgetID: "shared-1", PublicData: "new-state", CacheTime: 300}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetUniqueGift(ctx, GetUniqueGiftParams{Name: "gift-name"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetChatShareThemeBackdrops(ctx); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteBusinessAccountGift(ctx, DeleteBusinessAccountGiftParams{BusinessConnectionID: "business-1", OwnedGiftID: "owned-1"}); err != nil {
		t.Fatal(err)
	}
	entities := []MessageEntity{{Type: "bold", Offset: 0, Length: 4}}
	if err := client.VerifyUser(ctx, VerifyUserParams{UserID: 8, CustomDescription: "Verified", CustomDescriptionParseMode: "HTML", CustomDescriptionEntities: entities}); err != nil {
		t.Fatal(err)
	}
	if err := client.VerifyChat(ctx, VerifyChatParams{ChatID: 7, CustomDescription: "Verified", CustomDescriptionParseMode: "HTML", CustomDescriptionEntities: entities}); err != nil {
		t.Fatal(err)
	}
}

func TestEditSharedWidgetDataAllowsEmptyValueAndServerClampedCacheTime(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body := make(map[string]any)
		rawBody, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		if len(rawBody) != 0 {
			if err := json.Unmarshal(rawBody, &body); err != nil {
				t.Fatal(err)
			}
		}
		if body["widget_id"] != "shared-1" || body["public_data"] != "" || body["cache_time"] != float64(3601) {
			t.Fatalf("editSharedWidgetData body = %#v", body)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	if err := client.EditSharedWidgetData(context.Background(), EditSharedWidgetDataParams{
		WidgetID: "shared-1", PublicData: "", CacheTime: 3601,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestWidgetPayloadLimitsUseUTF8Bytes(t *testing.T) {
	t.Parallel()

	if err := validateWidgetState(strings.Repeat("é", 512), "state"); err != nil {
		t.Fatalf("1024-byte UTF-8 state rejected: %v", err)
	}
	if err := validateWidgetState(strings.Repeat("é", 513), "state"); err == nil {
		t.Fatal("expected state larger than 1024 UTF-8 bytes to fail")
	}
	if err := validateWidgetState("bad\x00state", "state"); err == nil {
		t.Fatal("expected NUL byte to fail")
	}
	if err := validateWidgetHTML("<main>valid</main>"); err != nil {
		t.Fatalf("valid widget HTML rejected: %v", err)
	}
	if err := validateWidgetHTML(" "); err != nil {
		t.Fatalf("non-empty whitespace HTML rejected: %v", err)
	}
	if err := validateWidgetHTML(""); err == nil {
		t.Fatal("expected empty HTML to fail")
	}
	if err := validateWidgetPayload("\x00", "answer data"); err != nil {
		t.Fatalf("NUL in answer data rejected: %v", err)
	}
	if err := validateWidgetHTML(strings.Repeat("x", maxWidgetHTMLBytes+1)); err == nil {
		t.Fatal("expected HTML larger than 32768 bytes to fail")
	}

	random := rand.New(rand.NewSource(104))
	data := make([]byte, 8192)
	for index := range data {
		data[index] = byte(32 + random.Intn(95))
	}
	if err := validateWidgetHTML(string(data)); err == nil {
		t.Fatal("expected HTML over the 4096-byte compressed limit to fail")
	}
}

func TestSendMessageEphemeralCallbackDoesNotRequireChatID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body := make(map[string]any)
		rawBody, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		if len(rawBody) != 0 {
			if err := json.Unmarshal(rawBody, &body); err != nil {
				t.Fatal(err)
			}
		}
		if _, exists := body["chat_id"]; exists {
			t.Fatalf("chat_id must be omitted for callback-scoped ephemeral send: %#v", body)
		}
		ephemeral, ok := body["ephemeral_message_parameters"].(map[string]any)
		if !ok || ephemeral["callback_query_id"] != "callback-1" {
			t.Fatalf("ephemeral_message_parameters = %#v", body["ephemeral_message_parameters"])
		}
		if _, exists := ephemeral["receiver_user_id"]; exists {
			t.Fatalf("receiver_user_id must remain optional: %#v", ephemeral)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"message_id":3,"chat":{"id":7,"type":"private"}}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	_, err := client.SendMessage(context.Background(), SendMessageParams{
		Text:                       "ephemeral",
		EphemeralMessageParameters: &EphemeralMessageParameters{CallbackQueryID: "callback-1"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSendPhotoUploadEphemeralCallbackDoesNotRequireChatID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if request.FormValue("chat_id") != "" {
			t.Fatalf("chat_id must be omitted, got %q", request.FormValue("chat_id"))
		}
		var ephemeral map[string]any
		if err := json.Unmarshal([]byte(request.FormValue("ephemeral_message_parameters")), &ephemeral); err != nil {
			t.Fatal(err)
		}
		if ephemeral["callback_query_id"] != "callback-2" {
			t.Fatalf("ephemeral parameters = %#v", ephemeral)
		}
		if _, exists := ephemeral["receiver_user_id"]; exists {
			t.Fatalf("receiver_user_id should be omitted: %#v", ephemeral)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"message_id":4,"chat":{"id":7,"type":"private"}}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	_, err := client.SendPhotoUpload(context.Background(), SendPhotoParams{
		Photo:                      "attach://photo",
		EphemeralMessageParameters: &EphemeralMessageParameters{CallbackQueryID: "callback-2"},
	}, "photo.jpg", strings.NewReader("image"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestOptionalChatIDStillRequiresEphemeralCallback(t *testing.T) {
	t.Parallel()

	client := New("TOKEN")
	_, err := client.SendMessage(context.Background(), SendMessageParams{Text: "missing target"})
	if err == nil || !strings.Contains(err.Error(), "callback_query_id") {
		t.Fatalf("error = %v, want missing chat_id validation", err)
	}
}
