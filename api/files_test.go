package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDownloadFileStreamsBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/file/botTOKEN/photos/a.jpg" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte("content"))
	}))
	defer server.Close()

	bot := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	var output bytes.Buffer
	written, err := bot.DownloadFile(context.Background(), "photos/a.jpg", &output)
	if err != nil {
		t.Fatal(err)
	}
	if written != 7 || output.String() != "content" {
		t.Fatalf("download = %d %q", written, output.String())
	}
}

func TestDownloadRejectsTraversal(t *testing.T) {
	t.Parallel()

	bot := New("TOKEN")
	if _, err := bot.OpenFile(context.Background(), "../token"); err == nil {
		t.Fatal("expected invalid path error")
	}
}
