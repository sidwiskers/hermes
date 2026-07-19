package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendLivePhotoStreamsBothComponents(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if request.FormValue("live_photo") != "attach://clip" || request.FormValue("photo") != "attach://still" {
			t.Fatalf("form = %#v", request.MultipartForm.Value)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"message_id":1,"chat":{"id":7,"type":"private"}}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	_, err := client.SendLivePhotoUpload(context.Background(), SendLivePhotoParams{
		ChatID: 7, LivePhoto: Attachment("clip"), Photo: Attachment("still"),
	},
		NewUpload("clip", "clip.mp4", strings.NewReader("video")),
		NewUpload("still", "still.jpg", strings.NewReader("image")),
	)
	if err != nil {
		t.Fatal(err)
	}
}
