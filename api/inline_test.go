package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnswerInlineQueryContract(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		results := body["results"].([]any)
		result := results[0].(map[string]any)
		if result["type"] != "article" || result["id"] != "one" {
			t.Fatalf("result = %#v", result)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	err := client.AnswerInlineQuery(context.Background(), AnswerInlineQueryParams{
		InlineQueryID: "query",
		Results: []InlineQueryResult{InlineQueryResultArticle{
			ID:                  "one",
			Title:               "Hermes",
			InputMessageContent: InputTextMessageContent{MessageText: "Fast"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSavePreparedKeyboardButtonValidation(t *testing.T) {
	t.Parallel()

	client := New("TOKEN")
	_, err := client.SavePreparedKeyboardButton(context.Background(), SavePreparedKeyboardButtonParams{
		UserID: 1,
		Button: KeyboardButton{Text: "Choose", RequestUsers: &KeyboardButtonRequestUsers{RequestID: 1}, RequestChat: &KeyboardButtonRequestChat{RequestID: 2}},
	})
	if err == nil {
		t.Fatal("expected multiple-action validation error")
	}
}

func TestInlineQueryResultTypedNilIsRejected(t *testing.T) {
	t.Parallel()
	var result *InlineQueryResultArticle
	err := validateInlineQueryResult(result)
	if err == nil {
		t.Fatal("expected typed nil result to fail validation")
	}
}
