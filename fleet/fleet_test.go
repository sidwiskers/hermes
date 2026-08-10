package fleet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sidwiskers/hermes"
)

func TestMountValidationAndStatus(t *testing.T) {
	t.Parallel()

	var nilFleet *Fleet
	if err := nilFleet.Mount("bot", hermes.New("TOKEN")); !errors.Is(err, ErrFleetRequired) {
		t.Fatalf("nil Fleet error = %v", err)
	}
	host := New()
	if err := host.Mount("", hermes.New("TOKEN")); !errors.Is(err, ErrNameRequired) {
		t.Fatalf("empty name error = %v", err)
	}
	if err := host.Mount("bot", nil); !errors.Is(err, ErrBotRequired) {
		t.Fatalf("nil bot error = %v", err)
	}
	if err := host.Mount("bot", new(hermes.Bot)); !errors.Is(err, ErrBotRequired) {
		t.Fatalf("empty bot error = %v", err)
	}

	first := host.NewBot("FIRST", hermes.WithBotUsername("first"))
	if err := host.Mount("first", first); err != nil {
		t.Fatal(err)
	}
	if err := host.Mount("first", host.NewBot("OTHER")); !errors.Is(err, ErrDuplicateName) {
		t.Fatalf("duplicate name error = %v", err)
	}
	if err := host.Mount("alias", first); !errors.Is(err, ErrDuplicateBot) {
		t.Fatalf("duplicate bot error = %v", err)
	}
	if err := host.Mount("invalid", host.NewBot("INVALID"), WithWebhook("relative", hermes.WebhookOptions{})); !errors.Is(err, ErrWebhookPath) {
		t.Fatalf("invalid path error = %v", err)
	}
	if err := host.Mount("webhook", host.NewBot("WEBHOOK"), WithWebhook("/shared", hermes.WebhookOptions{})); err != nil {
		t.Fatal(err)
	}
	if err := host.Mount("duplicate-path", host.NewBot("DUPLICATE"), WithWebhook("/shared", hermes.WebhookOptions{})); !errors.Is(err, ErrDuplicateWebhookPath) {
		t.Fatalf("duplicate path error = %v", err)
	}

	if host.Len() != 2 {
		t.Fatalf("mounted bots = %d, want 2", host.Len())
	}
	statuses := host.Status()
	if len(statuses) != 2 || statuses[0].Name != "first" || statuses[0].Mode != ModePolling || statuses[0].State != StateRegistered {
		t.Fatalf("statuses = %#v", statuses)
	}
	if err := host.Run(context.Background()); !errors.Is(err, ErrWebhookAddressRequired) {
		t.Fatalf("missing webhook address error = %v", err)
	}
}

func TestRunRequiresMountedBots(t *testing.T) {
	t.Parallel()

	if err := New().Run(context.Background()); !errors.Is(err, ErrNoBots) {
		t.Fatalf("empty Fleet error = %v", err)
	}
	var host *Fleet
	if err := host.Run(context.Background()); !errors.Is(err, ErrFleetRequired) {
		t.Fatalf("nil Fleet error = %v", err)
	}
}

func TestFleetIsolatesPreparationFailure(t *testing.T) {
	t.Parallel()

	polled := make(chan struct{}, 1)
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch methodName(request) {
		case "getMe":
			return telegramResponse(request, http.StatusOK, map[string]any{
				"ok": false, "error_code": http.StatusUnauthorized, "description": "Unauthorized",
			}), nil
		case "getUpdates":
			select {
			case polled <- struct{}{}:
			default:
			}
			<-request.Context().Done()
			return nil, request.Context().Err()
		default:
			return telegramResponse(request, http.StatusOK, map[string]any{"ok": true, "result": true}), nil
		}
	})
	failures := make(chan *Failure, 2)
	host := New(
		WithHTTPClient(&http.Client{Transport: transport}),
		WithErrorHandler(func(failure *Failure) { failures <- failure }),
	)
	bad := host.NewBot("BAD")
	good := host.NewBot("GOOD", hermes.WithBotUsername("good_bot"))
	if err := host.Mount("bad", bad); err != nil {
		t.Fatal(err)
	}
	if err := host.Mount("good", good); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- host.Run(ctx) }()

	select {
	case failure := <-failures:
		if failure.Bot != "bad" || failure.Phase != PhasePrepare {
			t.Fatalf("failure = %#v", failure)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("preparation failure was not reported")
	}
	select {
	case <-polled:
	case <-time.After(2 * time.Second):
		t.Fatal("healthy bot did not continue polling")
	}
	statuses := host.Status()
	if statuses[0].State != StateFailed || statuses[1].State != StateRunning {
		t.Fatalf("statuses = %#v", statuses)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run error after cancellation = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Fleet did not stop")
	}
}

func TestFleetIsolatesWebhookListenerFailure(t *testing.T) {
	t.Parallel()

	polled := make(chan struct{}, 1)
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if methodName(request) == "getUpdates" {
			select {
			case polled <- struct{}{}:
			default:
			}
			<-request.Context().Done()
			return nil, request.Context().Err()
		}
		return telegramResponse(request, http.StatusOK, map[string]any{"ok": true, "result": true}), nil
	})
	failures := make(chan *Failure, 1)
	host := New(
		WithHTTPClient(&http.Client{Transport: transport}),
		WithWebhookAddress("127.0.0.1:not-a-port"),
		WithErrorHandler(func(failure *Failure) { failures <- failure }),
	)
	polling := host.NewBot("POLLING", hermes.WithBotUsername("polling_bot"))
	webhook := host.NewBot("WEBHOOK", hermes.WithBotUsername("webhook_bot"))
	if err := host.Mount("polling", polling); err != nil {
		t.Fatal(err)
	}
	if err := host.Mount("webhook", webhook, WithWebhook("/webhook", hermes.WebhookOptions{})); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- host.Run(ctx) }()
	select {
	case failure := <-failures:
		if failure.Bot != "" || failure.Phase != PhaseWebhookServer {
			t.Fatalf("failure = %#v", failure)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("webhook listener failure was not reported")
	}
	select {
	case <-polled:
	case <-time.After(2 * time.Second):
		t.Fatal("polling bot did not continue")
	}
	statuses := host.Status()
	if statuses[0].State != StateRunning || statuses[1].State != StateFailed {
		t.Fatalf("statuses = %#v", statuses)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run error after cancellation = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Fleet did not stop")
	}
}

func TestFleetStopAllAndErrorHandlerPanic(t *testing.T) {
	t.Parallel()

	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return telegramResponse(request, http.StatusOK, map[string]any{
			"ok": false, "error_code": http.StatusUnauthorized, "description": "Unauthorized",
		}), nil
	})
	host := New(
		WithHTTPClient(&http.Client{Transport: transport}),
		WithFailurePolicy(StopAllOnFailure),
		WithErrorHandler(func(*Failure) { panic("contained") }),
	)
	if err := host.Mount("bad", host.NewBot("BAD")); err != nil {
		t.Fatal(err)
	}
	err := host.Run(context.Background())
	var failure *Failure
	if !errors.As(err, &failure) || failure.Bot != "bad" || failure.Phase != PhasePrepare {
		t.Fatalf("Run error = %v", err)
	}
}

func TestFleetRejectsMutationWhileRunning(t *testing.T) {
	t.Parallel()

	polled := make(chan struct{}, 1)
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if methodName(request) == "getUpdates" {
			polled <- struct{}{}
			<-request.Context().Done()
			return nil, request.Context().Err()
		}
		return telegramResponse(request, http.StatusOK, map[string]any{"ok": true, "result": true}), nil
	})
	host := New(WithHTTPClient(&http.Client{Transport: transport}))
	if err := host.Mount("first", host.NewBot("FIRST", hermes.WithBotUsername("first"))); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- host.Run(ctx) }()
	select {
	case <-polled:
	case <-time.After(2 * time.Second):
		t.Fatal("bot did not poll")
	}
	if err := host.Mount("second", host.NewBot("SECOND")); !errors.Is(err, ErrRunning) {
		t.Fatalf("mount while running error = %v", err)
	}
	if err := host.Run(context.Background()); !errors.Is(err, ErrRunning) {
		t.Fatalf("second Run error = %v", err)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestFleetStatusIsRaceSafe(t *testing.T) {
	t.Parallel()

	polled := make(chan struct{}, 1)
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if methodName(request) == "getUpdates" {
			select {
			case polled <- struct{}{}:
			default:
			}
			<-request.Context().Done()
			return nil, request.Context().Err()
		}
		return telegramResponse(request, http.StatusOK, map[string]any{"ok": true, "result": true}), nil
	})
	host := New(WithHTTPClient(&http.Client{Transport: transport}))
	if err := host.Mount("bot", host.NewBot("TOKEN", hermes.WithBotUsername("bot"))); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- host.Run(ctx) }()
	<-polled

	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 1_000 {
				_ = host.Status()
				_ = host.WebhookAddress()
				_ = host.Len()
			}
		}()
	}
	wait.Wait()
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func methodName(request *http.Request) string {
	path := strings.TrimRight(request.URL.Path, "/")
	if index := strings.LastIndexByte(path, '/'); index >= 0 {
		return path[index+1:]
	}
	return path
}

func telegramResponse(request *http.Request, status int, body any) *http.Response {
	data, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(data)),
		Request:    request,
	}
}

func waitForAddress(t *testing.T, host *Fleet) string {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if address := host.WebhookAddress(); address != "" {
			return address
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("Fleet webhook server did not start")
	return ""
}

func postUpdate(t *testing.T, address, path, secret, body string) (*http.Response, []byte) {
	t.Helper()
	request, err := http.NewRequest(http.MethodPost, "http://"+address+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if secret != "" {
		request.Header.Set(hermes.WebhookSecretHeader, secret)
	}
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	return response, data
}

func messageUpdateJSON(id int64, text string) string {
	data, _ := json.Marshal(map[string]any{
		"update_id": id,
		"message": map[string]any{
			"message_id": id,
			"date":       1,
			"from": map[string]any{
				"id": id, "is_bot": false, "first_name": "User",
			},
			"chat": map[string]any{
				"id": id, "type": "private", "first_name": "User",
			},
			"text": text,
		},
	})
	return string(data)
}
