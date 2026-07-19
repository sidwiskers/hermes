package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUpdatesRawMode(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"result":[{"update_id":9,"future":{"x":1}}]}`))
	}))
	defer server.Close()

	fast := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	updates, err := fast.GetUpdates(context.Background(), GetUpdatesParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(updates) != 1 || len(updates[0].Raw) != 0 {
		t.Fatalf("fast updates = %#v", updates)
	}

	preserving := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()), WithRawUpdates(true))
	updates, err = preserving.GetUpdates(context.Background(), GetUpdatesParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(updates) != 1 || string(updates[0].Raw) != `{"update_id":9,"future":{"x":1}}` {
		t.Fatalf("preserved updates = %#v", updates)
	}
}

func TestRawUpdatesEnabled(t *testing.T) {
	t.Parallel()
	if New("TOKEN").RawUpdatesEnabled() {
		t.Fatal("raw updates enabled by default")
	}
	if !New("TOKEN", WithRawUpdates(true)).RawUpdatesEnabled() {
		t.Fatal("raw updates option ignored")
	}
}
