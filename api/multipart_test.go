package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendPhotoUploadStreamsMultipart(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("chat_id") != "99" || r.FormValue("photo") != "attach://photo" {
			t.Fatalf("unexpected fields: %#v", r.Form)
		}
		if r.FormValue("receiver_user_id") != "7" || r.FormValue("callback_query_id") != "cb" {
			t.Fatalf("missing ephemeral fields: %#v", r.Form)
		}

		file, header, err := r.FormFile("photo")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			t.Fatal(err)
		}
		if header.Filename != "image.txt" || string(content) != "image-bytes" {
			t.Fatalf("unexpected file: %q %q", header.Filename, content)
		}
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":0,"chat":{"id":99,"type":"supergroup"},"receiver_user":{"id":7,"is_bot":false,"first_name":"A"},"ephemeral_message_id":4}}`))
	}))
	defer server.Close()

	bot := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	message, err := bot.SendPhotoUpload(context.Background(), SendPhotoParams{
		ChatID:          int64(99),
		ReceiverUserID:  7,
		CallbackQueryID: "cb",
		Caption:         "caption",
	}, "image.txt", strings.NewReader("image-bytes"))
	if err != nil {
		t.Fatal(err)
	}
	if message.EphemeralMessageID != 4 {
		t.Fatalf("unexpected message: %#v", message)
	}
}
