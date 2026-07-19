package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type recordingObserver struct {
	mu       sync.Mutex
	started  []CallEvent
	finished []CallResult
	panic    bool
}

func (o *recordingObserver) StartCall(ctx context.Context, event CallEvent) context.Context {
	if o.panic {
		panic("observer start")
	}
	o.mu.Lock()
	o.started = append(o.started, event)
	o.mu.Unlock()
	return ctx
}

func (o *recordingObserver) FinishCall(_ context.Context, _ CallEvent, result CallResult) {
	if o.panic {
		panic("observer finish")
	}
	o.mu.Lock()
	o.finished = append(o.finished, result)
	o.mu.Unlock()
}

func TestObserverSeesTelegramEnvelopeResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/botTOKEN/getMe" {
			t.Fatalf("path=%s", request.URL.Path)
		}
		_, _ = writer.Write([]byte(`{"ok":false,"error_code":429,"description":"slow down"}`))
	}))
	defer server.Close()
	observer := new(recordingObserver)
	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()), WithObserver(observer))
	err := client.Call(context.Background(), "getMe", nil, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error=%v", err)
	}
	if len(observer.started) != 1 || observer.started[0] != (CallEvent{Method: "getMe", Kind: CallJSON}) {
		t.Fatalf("started=%+v", observer.started)
	}
	if len(observer.finished) != 1 || !errors.As(observer.finished[0].Err, &apiErr) || observer.finished[0].Duration < 0 {
		t.Fatalf("finished=%+v", observer.finished)
	}
}

func TestObserverTracksDownloadThroughEOF(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte("payload"))
	}))
	defer server.Close()
	observer := new(recordingObserver)
	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()), WithObserver(observer))
	var destination bytes.Buffer
	if _, err := client.DownloadFile(context.Background(), "documents/a.txt", &destination); err != nil {
		t.Fatal(err)
	}
	if len(observer.started) != 1 || observer.started[0] != (CallEvent{Method: "downloadFile", Kind: CallDownload}) {
		t.Fatalf("started=%+v", observer.started)
	}
	if len(observer.finished) != 1 || observer.finished[0].Err != nil {
		t.Fatalf("finished=%+v", observer.finished)
	}
}

func TestObserverContextAndPanics(t *testing.T) {
	observer := &recordingObserver{panic: true}
	client := New("TOKEN", WithObserver(observer), WithHTTPClient(&http.Client{Transport: roundTripFunc(
		func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       http.NoBody,
			}, nil
		},
	)}))
	if err := client.Call(context.Background(), "getMe", nil, nil); err == nil {
		t.Fatal("expected malformed-response error, not observer panic")
	}
}
