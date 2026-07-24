package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAllPrimaryMediaUploadsUseExpectedField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		method string
		field  string
		call   func(context.Context, *Client, io.Reader) error
	}{
		{"sendPhoto", "photo", func(ctx context.Context, b *Client, r io.Reader) error {
			_, err := b.SendPhotoUpload(ctx, SendPhotoParams{ChatID: 1}, "x.bin", r)
			return err
		}},
		{"sendAnimation", "animation", func(ctx context.Context, b *Client, r io.Reader) error {
			_, err := b.SendAnimationUpload(ctx, SendAnimationParams{ChatID: 1}, "x.bin", r)
			return err
		}},
		{"sendAudio", "audio", func(ctx context.Context, b *Client, r io.Reader) error {
			_, err := b.SendAudioUpload(ctx, SendAudioParams{ChatID: 1}, "x.bin", r)
			return err
		}},
		{"sendDocument", "document", func(ctx context.Context, b *Client, r io.Reader) error {
			_, err := b.SendDocumentUpload(ctx, SendDocumentParams{ChatID: 1}, "x.bin", r)
			return err
		}},
		{"sendSticker", "sticker", func(ctx context.Context, b *Client, r io.Reader) error {
			_, err := b.SendStickerUpload(ctx, SendStickerParams{ChatID: 1}, "x.bin", r)
			return err
		}},
		{"sendVideo", "video", func(ctx context.Context, b *Client, r io.Reader) error {
			_, err := b.SendVideoUpload(ctx, SendVideoParams{ChatID: 1}, "x.bin", r)
			return err
		}},
		{"sendVideoNote", "video_note", func(ctx context.Context, b *Client, r io.Reader) error {
			_, err := b.SendVideoNoteUpload(ctx, SendVideoNoteParams{ChatID: 1}, "x.bin", r)
			return err
		}},
		{"sendVoice", "voice", func(ctx context.Context, b *Client, r io.Reader) error {
			_, err := b.SendVoiceUpload(ctx, SendVoiceParams{ChatID: 1}, "x.bin", r)
			return err
		}},
	}

	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseMultipartForm(1 << 20); err != nil {
					t.Fatal(err)
				}
				if r.FormValue(test.field) != "attach://"+test.field {
					t.Fatalf("field %s = %q", test.field, r.FormValue(test.field))
				}
				file, header, err := r.FormFile(test.field)
				if err != nil {
					t.Fatal(err)
				}
				defer file.Close()
				data, _ := io.ReadAll(file)
				if header.Filename != "x.bin" || string(data) != "bytes" {
					t.Fatalf("upload = %q %q", header.Filename, data)
				}
				_, _ = io.WriteString(w, `{"ok":true,"result":{"message_id":1,"chat":{"id":1,"type":"private"}}}`)
			}))
			defer server.Close()
			bot := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
			if err := test.call(context.Background(), bot, strings.NewReader("bytes")); err != nil {
				t.Fatal(err)
			}
		})
	}
}
