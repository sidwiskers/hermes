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
		{"sendLivePhoto", func(ctx context.Context, b *Client) error {
			_, err := b.SendLivePhoto(ctx, SendLivePhotoParams{ChatID: 1, LivePhoto: "live-id", Photo: "photo-id", ReceiverUserID: 2, CallbackQueryID: "cb"})
			return err
		}},
		{"sendRichMessage", func(ctx context.Context, b *Client) error {
			_, err := b.SendRichMessage(ctx, SendRichMessageParams{
				ChatID:                     1,
				RichMessage:                InputRichMessage{HTML: "<b>x</b>"},
				EphemeralMessageParameters: &EphemeralMessageParameters{ReceiverUserID: 2, CallbackQueryID: "cb"},
			})
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
				ephemeral, ok := params["ephemeral_message_parameters"].(map[string]any)
				if !ok || ephemeral["receiver_user_id"] != float64(2) || ephemeral["callback_query_id"] != "cb" {
					t.Fatalf("ephemeral fields missing: %#v", params)
				}
				if _, exists := params["receiver_user_id"]; exists {
					t.Fatalf("legacy receiver_user_id leaked to wire: %#v", params)
				}
				if _, exists := params["callback_query_id"]; exists {
					t.Fatalf("legacy callback_query_id leaked to wire: %#v", params)
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

func TestEphemeralReplacementParameters(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		parameters := body["ephemeral_message_parameters"].(map[string]any)
		if parameters["replace_callback_query_message"] != true {
			t.Fatalf("ephemeral parameters = %#v", parameters)
		}
		_, _ = fmt.Fprint(writer, `{"ok":true,"result":{"message_id":0,"chat":{"id":1,"type":"supergroup"},"ephemeral_message_id":9}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	_, err := client.SendMessage(context.Background(), SendMessageParams{
		ChatID: 1,
		Text:   "replacement",
		EphemeralMessageParameters: &EphemeralMessageParameters{
			ReceiverUserID:              2,
			CallbackQueryID:             "cb",
			ReplaceCallbackQueryMessage: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}
