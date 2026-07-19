package hermes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestWebhookReplyHandlerRoutesSynchronously(t *testing.T) {
	t.Parallel()
	bot := New("TOKEN", WithBotUsername("bot"), WithMaxConcurrentUpdates(1))
	bot.Command("start", func(c *Context) error {
		chatID, _ := c.ChatID()
		return c.RespondWebhook("sendMessage", SendMessageParams{ChatID: chatID, Text: "hello"})
	})
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"update_id":2,"message":{"message_id":1,"from":{"id":7,"is_bot":false,"first_name":"Ada"},"chat":{"id":9,"type":"private"},"text":"/start"}}`,
	))
	response := httptest.NewRecorder()
	bot.WebhookReplyHandler(WebhookOptions{}).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["method"] != "sendMessage" || body["chat_id"] != float64(9) || body["text"] != "hello" {
		t.Fatalf("body=%v", body)
	}
	bot.Wait()
}

func TestWebhookReplySurvivesDerivedMiddlewareContext(t *testing.T) {
	t.Parallel()
	bot := New("TOKEN", WithBotUsername("bot"))
	bot.Use(func(next Handler) Handler {
		return func(c *Context) error {
			derived := *c
			derived.Context = context.WithValue(c.Context, struct{}{}, "trace")
			return next(&derived)
		}
	})
	bot.OnUpdate(func(c *Context) error {
		return c.RespondWebhook("sendMessage", SendMessageParams{ChatID: int64(9), Text: "observed"})
	})

	response := httptest.NewRecorder()
	bot.WebhookReplyHandler(WebhookOptions{}).ServeHTTP(response,
		httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"update_id":1}`)))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"method":"sendMessage"`) {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
}

func TestWebhookReplyHandlerReportsErrorAndPanic(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		handler Handler
	}{
		{name: "error", handler: func(*Context) error { return errors.New("failed") }},
		{name: "panic", handler: func(*Context) error { panic("failed") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			reported := make(chan error, 1)
			bot := New("TOKEN", WithBotUsername("bot"), WithErrorHandler(func(_ *Context, err error) { reported <- err }))
			bot.OnUpdate(test.handler)
			response := httptest.NewRecorder()
			bot.WebhookReplyHandler(WebhookOptions{}).ServeHTTP(response,
				httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"update_id":1}`)))
			if response.Code != http.StatusInternalServerError {
				t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
			}
			select {
			case err := <-reported:
				if err == nil {
					t.Fatal("nil reported error")
				}
			case <-time.After(time.Second):
				t.Fatal("error not reported")
			}
		})
	}
}

func TestWebhookReplyHandlerSharesConcurrencyBound(t *testing.T) {
	t.Parallel()
	started := make(chan struct{})
	release := make(chan struct{})
	bot := New("TOKEN", WithBotUsername("bot"), WithMaxConcurrentUpdates(1))
	bot.OnUpdate(func(*Context) error {
		close(started)
		<-release
		return nil
	})
	handler := bot.WebhookReplyHandler(WebhookOptions{})
	firstDone := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"update_id":1}`)))
		firstDone <- response
	}()
	<-started
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"update_id":2}`)))
	if second.Code != http.StatusServiceUnavailable || second.Header().Get("Retry-After") != "1" {
		t.Fatalf("status=%d retry=%q", second.Code, second.Header().Get("Retry-After"))
	}
	close(release)
	if first := <-firstDone; first.Code != http.StatusOK {
		t.Fatalf("first status=%d", first.Code)
	}
	bot.Wait()
}

func TestWebhookSecretAndDispatch(t *testing.T) {
	t.Parallel()

	bot := New("TOKEN", WithMaxConcurrentUpdates(1))
	var wg sync.WaitGroup
	wg.Add(1)
	bot.Command("start", func(c *Context) error {
		defer wg.Done()
		if c.Sender() == nil || c.Sender().ID != 7 {
			t.Fatalf("unexpected sender: %#v", c.Sender())
		}
		return nil
	})

	handler := bot.WebhookHandler(WebhookOptions{Secret: "secret"})

	unauthorized := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"update_id":1}`))
	unauthorizedResponse := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedResponse, unauthorized)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", unauthorizedResponse.Code)
	}

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"update_id":2,"message":{"message_id":1,"from":{"id":7,"is_bot":false,"first_name":"Ada"},"chat":{"id":9,"type":"private"},"text":"/start"}}`,
	))
	request.Header.Set(webhookSecretHeader, "secret")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handler did not run")
	}
	bot.Wait()
}

func TestWebhookQueueBackpressure(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	started := make(chan struct{})
	bot := New("TOKEN", WithMaxConcurrentUpdates(1))
	bot.OnUpdate(func(*Context) error {
		close(started)
		<-release
		return nil
	})
	handler := bot.WebhookHandler(WebhookOptions{})

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"update_id":1}`)))
	if first.Code != http.StatusOK {
		t.Fatalf("first status = %d", first.Code)
	}
	<-started

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"update_id":2}`)))
	if second.Code != http.StatusServiceUnavailable {
		t.Fatalf("second status = %d", second.Code)
	}

	close(release)
	bot.Wait()
}

func TestHandleNilUpdate(t *testing.T) {
	t.Parallel()
	if err := New("TOKEN").Handle(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}
