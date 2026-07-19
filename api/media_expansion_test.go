package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEphemeralParametersAcrossSupportedSendMethods(t *testing.T) {
	t.Parallel()

	methods := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{"sendMessage", func(ctx context.Context, b *Client) error {
			_, err := b.SendMessage(ctx, SendMessageParams{ChatID: 1, Text: "x", ReceiverUserID: 2, CallbackQueryID: "cb"})
			return err
		}},
		{"sendPhoto", func(ctx context.Context, b *Client) error {
			_, err := b.SendPhoto(ctx, SendPhotoParams{ChatID: 1, Photo: "id", ReceiverUserID: 2, CallbackQueryID: "cb"})
			return err
		}},
		{"sendAnimation", func(ctx context.Context, b *Client) error {
			_, err := b.SendAnimation(ctx, SendAnimationParams{ChatID: 1, Animation: "id", ReceiverUserID: 2, CallbackQueryID: "cb"})
			return err
		}},
		{"sendAudio", func(ctx context.Context, b *Client) error {
			_, err := b.SendAudio(ctx, SendAudioParams{ChatID: 1, Audio: "id", ReceiverUserID: 2, CallbackQueryID: "cb"})
			return err
		}},
		{"sendDocument", func(ctx context.Context, b *Client) error {
			_, err := b.SendDocument(ctx, SendDocumentParams{ChatID: 1, Document: "id", ReceiverUserID: 2, CallbackQueryID: "cb"})
			return err
		}},
		{"sendSticker", func(ctx context.Context, b *Client) error {
			_, err := b.SendSticker(ctx, SendStickerParams{ChatID: 1, Sticker: "id", ReceiverUserID: 2, CallbackQueryID: "cb"})
			return err
		}},
		{"sendVideo", func(ctx context.Context, b *Client) error {
			_, err := b.SendVideo(ctx, SendVideoParams{ChatID: 1, Video: "id", ReceiverUserID: 2, CallbackQueryID: "cb"})
			return err
		}},
		{"sendVideoNote", func(ctx context.Context, b *Client) error {
			_, err := b.SendVideoNote(ctx, SendVideoNoteParams{ChatID: 1, VideoNote: "id", ReceiverUserID: 2, CallbackQueryID: "cb"})
			return err
		}},
		{"sendVoice", func(ctx context.Context, b *Client) error {
			_, err := b.SendVoice(ctx, SendVoiceParams{ChatID: 1, Voice: "id", ReceiverUserID: 2, CallbackQueryID: "cb"})
			return err
		}},
		{"sendContact", func(ctx context.Context, b *Client) error {
			_, err := b.SendContact(ctx, SendContactParams{ChatID: 1, PhoneNumber: "+1", FirstName: "A", ReceiverUserID: 2, CallbackQueryID: "cb"})
			return err
		}},
		{"sendLocation", func(ctx context.Context, b *Client) error {
			_, err := b.SendLocation(ctx, SendLocationParams{ChatID: 1, Latitude: 1, Longitude: 2, ReceiverUserID: 2, CallbackQueryID: "cb"})
			return err
		}},
		{"sendVenue", func(ctx context.Context, b *Client) error {
			_, err := b.SendVenue(ctx, SendVenueParams{ChatID: 1, Latitude: 1, Longitude: 2, Title: "T", Address: "A", ReceiverUserID: 2, CallbackQueryID: "cb"})
			return err
		}},
	}

	for _, test := range methods {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasSuffix(r.URL.Path, "/"+test.name) {
					t.Fatalf("path = %s", r.URL.Path)
				}
				var params map[string]any
				if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
					t.Fatal(err)
				}
				if params["receiver_user_id"] != float64(2) || params["callback_query_id"] != "cb" {
					t.Fatalf("ephemeral fields missing: %#v", params)
				}
				_, _ = fmt.Fprint(w, `{"ok":true,"result":{"message_id":0,"chat":{"id":1,"type":"supergroup"},"receiver_user":{"id":2,"is_bot":false,"first_name":"A"},"ephemeral_message_id":9}}`)
			}))
			defer server.Close()
			bot := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
			if err := test.call(context.Background(), bot); err != nil {
				t.Fatal(err)
			}
		})
	}
}
