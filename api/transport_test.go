package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSendMessageRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/botTOKEN/sendMessage" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("unexpected content type: %s", got)
		}

		var params map[string]any
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatal(err)
		}
		if params["chat_id"].(float64) != 99 || params["text"].(string) != "private" {
			t.Fatalf("unexpected params: %#v", params)
		}
		if params["receiver_user_id"].(float64) != 42 || params["callback_query_id"].(string) != "callback" {
			t.Fatalf("missing ephemeral params: %#v", params)
		}

		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":0,"chat":{"id":99,"type":"supergroup"},"receiver_user":{"id":42,"is_bot":false,"first_name":"Ada"},"ephemeral_message_id":7,"text":"private"}}`))
	}))
	defer server.Close()

	bot := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	message, err := bot.SendMessage(context.Background(), SendMessageParams{
		ChatID:          int64(99),
		Text:            "private",
		ReceiverUserID:  42,
		CallbackQueryID: "callback",
	})
	if err != nil {
		t.Fatal(err)
	}
	if message.EphemeralMessageID != 7 || message.ReceiverUser == nil || message.ReceiverUser.ID != 42 {
		t.Fatalf("unexpected message: %#v", message)
	}
}

func TestAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"ok":false,"error_code":429,"description":"Too Many Requests","parameters":{"retry_after":3}}`))
	}))
	defer server.Close()

	bot := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	err := bot.Call(context.Background(), "sendMessage", map[string]any{"chat_id": 1, "text": "x"}, nil)

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Code != 429 || apiErr.RetryAfter() != 3 {
		t.Fatalf("unexpected APIError: %#v", apiErr)
	}
}

func TestResponseLimit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"result":"` + strings.Repeat("x", 128) + `"}`))
	}))
	defer server.Close()

	bot := New(
		"TOKEN",
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
		WithResponseLimit(32),
	)
	err := bot.Call(context.Background(), "getMe", nil, nil)
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("expected response limit error, got %v", err)
	}
}

func TestMaximumResponseLimitDoesNotOverflow(t *testing.T) {
	t.Parallel()

	client := New(
		"TOKEN",
		WithResponseLimit(1<<63-1),
		WithHTTPClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return testResponse(http.StatusOK, `{"ok":true,"result":true}`), nil
		})}),
	)
	var result bool
	if err := client.Call(context.Background(), "getMe", nil, &result); err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("successful result was not decoded")
	}
}

func TestInvalidMethodRejected(t *testing.T) {
	t.Parallel()

	bot := New("TOKEN")
	err := bot.Call(context.Background(), "../getMe", nil, nil)
	if !errors.Is(err, ErrInvalidMethod) {
		t.Fatalf("expected ErrInvalidMethod, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestTransportErrorDoesNotLeakToken(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return nil, &url.Error{
			Op:  "Post",
			URL: request.URL.String(),
			Err: context.Canceled,
		}
	})}
	bot := New("SUPER_SECRET_TOKEN", WithHTTPClient(client))

	err := bot.Call(context.Background(), "getMe", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "SUPER_SECRET_TOKEN") {
		t.Fatalf("token leaked in error: %v", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("lost wrapped cause: %v", err)
	}
}
