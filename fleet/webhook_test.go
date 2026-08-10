package fleet

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sidwiskers/hermes"
)

func TestFleetSharesExactWebhookServer(t *testing.T) {
	t.Parallel()

	host := New(WithWebhookAddress("127.0.0.1:0"))
	alpha := host.NewBot("ALPHA", hermes.WithBotUsername("alpha"))
	beta := host.NewBot("BETA", hermes.WithBotUsername("beta"))
	var alphaCalls atomic.Int64
	var betaCalls atomic.Int64
	alpha.OnUpdate(func(*hermes.Context) error { alphaCalls.Add(1); return nil })
	beta.OnUpdate(func(*hermes.Context) error { betaCalls.Add(1); return nil })
	if err := host.Mount("alpha", alpha, WithWebhook("/alpha", hermes.WebhookOptions{Secret: "alpha-secret"})); err != nil {
		t.Fatal(err)
	}
	if err := host.Mount("beta", beta, WithWebhook("/beta", hermes.WebhookOptions{Secret: "beta-secret"})); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- host.Run(ctx) }()
	address := waitForAddress(t, host)

	response, _ := postUpdate(t, address, "/alpha", "alpha-secret", messageUpdateJSON(1, "alpha"))
	if response.StatusCode != http.StatusOK {
		t.Fatalf("alpha status = %d", response.StatusCode)
	}
	response, _ = postUpdate(t, address, "/beta", "beta-secret", messageUpdateJSON(2, "beta"))
	if response.StatusCode != http.StatusOK {
		t.Fatalf("beta status = %d", response.StatusCode)
	}
	response, _ = postUpdate(t, address, "/alpha/", "alpha-secret", messageUpdateJSON(3, "wrong path"))
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("inexact path status = %d", response.StatusCode)
	}
	response, _ = postUpdate(t, address, "/alpha", "wrong-secret", messageUpdateJSON(4, "unauthorized"))
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong secret status = %d", response.StatusCode)
	}
	alpha.Wait()
	beta.Wait()
	if alphaCalls.Load() != 1 || betaCalls.Load() != 1 {
		t.Fatalf("calls: alpha=%d beta=%d", alphaCalls.Load(), betaCalls.Load())
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Fleet did not stop")
	}
}

func TestFleetWebhookReplies(t *testing.T) {
	t.Parallel()

	host := New(WithWebhookAddress("127.0.0.1:0"))
	bot := host.NewBot("TOKEN", hermes.WithBotUsername("reply_bot"))
	bot.OnUpdate(func(c *hermes.Context) error {
		return c.RespondWebhook("sendMessage", hermes.SendMessageParams{ChatID: int64(7), Text: "direct"})
	})
	if err := host.Mount("reply", bot, WithWebhookReplies("/reply", hermes.WebhookOptions{Secret: "secret"})); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- host.Run(ctx) }()
	address := waitForAddress(t, host)

	response, body := postUpdate(t, address, "/reply", "secret", messageUpdateJSON(7, "reply"))
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.StatusCode, body)
	}
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	if result["method"] != "sendMessage" || result["text"] != "direct" {
		t.Fatalf("reply = %#v", result)
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestFleetWebhookShutdownDrainsHandlers(t *testing.T) {
	t.Parallel()

	host := New(WithWebhookAddress("127.0.0.1:0"))
	bot := host.NewBot("TOKEN", hermes.WithBotUsername("drain_bot"))
	started := make(chan struct{})
	release := make(chan struct{})
	bot.OnUpdate(func(*hermes.Context) error {
		close(started)
		<-release
		return nil
	})
	if err := host.Mount("drain", bot, WithWebhook("/drain", hermes.WebhookOptions{})); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- host.Run(ctx) }()
	address := waitForAddress(t, host)
	response, _ := postUpdate(t, address, "/drain", "", messageUpdateJSON(1, "block"))
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	<-started
	cancel()
	select {
	case err := <-done:
		t.Fatalf("Fleet stopped before handler drained: %v", err)
	case <-time.After(25 * time.Millisecond):
	}
	close(release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Fleet did not drain")
	}
}

func TestExactWebhookMux(t *testing.T) {
	t.Parallel()

	called := false
	handler := exactWebhookMux{
		"/exact": http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }),
	}
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/exact", nil))
	if !called {
		t.Fatal("exact route was not called")
	}
	called = false
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/exact/", nil))
	if called || response.Code != http.StatusNotFound {
		t.Fatalf("inexact route: called=%v status=%d", called, response.Code)
	}
}

type benchmarkWriter struct {
	header http.Header
}

func (w *benchmarkWriter) Header() http.Header          { return w.header }
func (*benchmarkWriter) Write(data []byte) (int, error) { return len(data), nil }
func (*benchmarkWriter) WriteHeader(int)                {}

func BenchmarkExactWebhookMux(b *testing.B) {
	handler := exactWebhookMux{
		"/bot": http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
	}
	request := httptest.NewRequest(http.MethodPost, "/bot", nil)
	writer := &benchmarkWriter{header: make(http.Header)}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		handler.ServeHTTP(writer, request)
	}
}
