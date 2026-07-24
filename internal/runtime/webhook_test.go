package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	telegram "github.com/sidwiskers/hermes/types"
)

func TestWebhookReplyHandlerEncodesMethodAndParameters(t *testing.T) {
	t.Parallel()
	handler := WebhookReplyHandler(WebhookOptions{}, func(_ context.Context, update *telegram.Update) (WebhookReply, error) {
		if update.UpdateID != 4 {
			t.Fatalf("update=%+v", update)
		}
		return WebhookReply{Method: "sendMessage", Params: struct {
			ChatID int64  `json:"chat_id"`
			Text   string `json:"text"`
		}{ChatID: 9, Text: "hello"}}, nil
	})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, webhookRequest(`{"update_id":4}`))
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("status=%d content-type=%q body=%s", response.Code, response.Header().Get("Content-Type"), response.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["method"] != "sendMessage" || body["chat_id"] != float64(9) || body["text"] != "hello" {
		t.Fatalf("body=%v", body)
	}
}

func TestWebhookReplyHandlerFailureClasses(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name       string
		handle     func(context.Context, *telegram.Update) (WebhookReply, error)
		options    WebhookOptions
		wantStatus int
	}{
		{
			name: "overload",
			handle: func(context.Context, *telegram.Update) (WebhookReply, error) {
				return WebhookReply{}, ErrQueueFull
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "handler error",
			handle: func(context.Context, *telegram.Update) (WebhookReply, error) {
				return WebhookReply{}, errors.New("sensitive internal error")
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "invalid parameters",
			handle: func(context.Context, *telegram.Update) (WebhookReply, error) {
				return WebhookReply{Method: "sendMessage", Params: []string{"not", "object"}}, nil
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "bounded response",
			handle: func(context.Context, *telegram.Update) (WebhookReply, error) {
				return WebhookReply{Method: "sendMessage", Params: map[string]string{"text": "long"}}, nil
			},
			options:    WebhookOptions{MaxResponseBytes: 8},
			wantStatus: http.StatusInternalServerError,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			WebhookReplyHandler(test.options, test.handle).ServeHTTP(response, webhookRequest(`{"update_id":1}`))
			if response.Code != test.wantStatus {
				t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
			}
			if strings.Contains(response.Body.String(), "sensitive") {
				t.Fatalf("handler error leaked: %q", response.Body.String())
			}
			if test.name == "overload" && response.Header().Get("Retry-After") != "1" {
				t.Fatalf("retry-after=%q", response.Header().Get("Retry-After"))
			}
		})
	}
}

func TestWebhookReplyHandlerEmptyResponse(t *testing.T) {
	t.Parallel()
	response := httptest.NewRecorder()
	WebhookReplyHandler(WebhookOptions{}, func(context.Context, *telegram.Update) (WebhookReply, error) {
		return WebhookReply{}, nil
	}).ServeHTTP(response, webhookRequest(`{"update_id":1}`))
	if response.Code != http.StatusOK || response.Body.Len() != 0 {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
}

func webhookRequest(body string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/telegram", strings.NewReader(body))
}

func TestWebhookFastAndRawModes(t *testing.T) {
	t.Parallel()

	payload := `{"update_id":4,"future":{"enabled":true}}`
	for _, test := range []struct {
		name     string
		preserve bool
	}{
		{name: "fast"},
		{name: "preserve", preserve: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var got *telegram.Update
			handler := WebhookHandler(WebhookOptions{PreserveRawUpdate: test.preserve}, func(_ context.Context, update *telegram.Update, wait bool) bool {
				if wait {
					t.Fatal("webhook enqueue should be non-blocking")
				}
				got = update
				return true
			})
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, webhookRequest(payload))
			if response.Code != http.StatusOK || got == nil || got.UpdateID != 4 {
				t.Fatalf("status=%d update=%#v", response.Code, got)
			}
			if test.preserve != (len(got.Raw) != 0) {
				t.Fatalf("preserve=%v raw=%q", test.preserve, got.Raw)
			}
		})
	}
}

func TestWebhookSecurityAndBackpressure(t *testing.T) {
	t.Parallel()

	handler := WebhookHandler(WebhookOptions{Secret: "secret"}, func(context.Context, *telegram.Update, bool) bool { return false })

	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, webhookRequest(`{"update_id":1}`))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", unauthorized.Code)
	}

	request := webhookRequest(`{"update_id":1}`)
	request.Header.Set(SecretHeader, "secret")
	full := httptest.NewRecorder()
	handler.ServeHTTP(full, request)
	if full.Code != http.StatusServiceUnavailable || full.Header().Get("Retry-After") != "1" {
		t.Fatalf("full status=%d retry=%q", full.Code, full.Header().Get("Retry-After"))
	}
}

func TestWebhookRejectsWrongMethodBeforeDispatch(t *testing.T) {
	t.Parallel()

	called := false
	handler := WebhookHandler(WebhookOptions{}, func(context.Context, *telegram.Update, bool) bool {
		called = true
		return true
	})
	request := httptest.NewRequest(http.MethodGet, "/telegram", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed || response.Header().Get("Allow") != http.MethodPost {
		t.Fatalf("status=%d allow=%q", response.Code, response.Header().Get("Allow"))
	}
	if called {
		t.Fatal("wrong-method request reached dispatch")
	}
}

func TestWebhookDispatchContextSurvivesRequestCancellation(t *testing.T) {
	t.Parallel()

	var dispatchErr error
	handler := WebhookHandler(WebhookOptions{}, func(ctx context.Context, _ *telegram.Update, _ bool) bool {
		dispatchErr = ctx.Err()
		return true
	})
	requestCtx, cancel := context.WithCancel(context.Background())
	cancel()
	request := webhookRequest(`{"update_id":1}`).WithContext(requestCtx)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || dispatchErr != nil {
		t.Fatalf("status=%d dispatch error=%v", response.Code, dispatchErr)
	}
}

func TestWebhookRejectsTrailingDataAndLargeBodies(t *testing.T) {
	t.Parallel()

	handler := WebhookHandler(WebhookOptions{MaxBodyBytes: 16}, func(context.Context, *telegram.Update, bool) bool { return true })

	trailing := httptest.NewRecorder()
	handler.ServeHTTP(trailing, webhookRequest(`{"update_id":1} {}`))
	if trailing.Code != http.StatusBadRequest && trailing.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("trailing status = %d", trailing.Code)
	}

	large := httptest.NewRecorder()
	handler.ServeHTTP(large, webhookRequest(`{"update_id":12345678901234567890}`))
	if large.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("large status = %d", large.Code)
	}
}

func TestServeWebhookValidatesInputs(t *testing.T) {
	t.Parallel()

	if err := ServeWebhook(context.Background(), "127.0.0.1:0", "/hook", nil, nil); !errors.Is(err, ErrWebhookHandlerRequired) {
		t.Fatalf("handler error = %v", err)
	}
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	for _, path := range []string{"hook", "/hook?query", "/{wildcard}"} {
		if err := ServeWebhook(context.Background(), "127.0.0.1:0", path, handler, nil); err == nil {
			t.Fatalf("path %q was accepted", path)
		}
	}
}

func TestServeWebhookDrainsOnListenFailure(t *testing.T) {
	t.Parallel()

	for _, ctx := range []context.Context{context.Background(), nil} {
		waited := false
		err := ServeWebhook(
			ctx,
			"127.0.0.1:not-a-port",
			"/hook",
			http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
			func() { waited = true },
		)
		if err == nil {
			t.Fatal("expected listen error")
		}
		if !waited {
			t.Fatal("dispatcher was not drained")
		}
	}
}

func TestServeWebhookCanceledContextShutsDown(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	waited := false
	err := ServeWebhook(
		ctx,
		"127.0.0.1:0",
		"/hook",
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		func() { waited = true },
	)
	if err != nil {
		t.Fatal(err)
	}
	if !waited {
		t.Fatal("dispatcher was not drained")
	}
}
