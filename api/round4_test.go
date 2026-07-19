package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendMediaGroupJSONContract(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/sendMediaGroup") {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		media, ok := body["media"].([]any)
		if !ok || len(media) != 2 {
			t.Fatalf("media = %#v", body["media"])
		}
		first := media[0].(map[string]any)
		second := media[1].(map[string]any)
		if first["type"] != "photo" || first["media"] != "photo-id" {
			t.Fatalf("first = %#v", first)
		}
		if second["type"] != "video" || second["media"] != "video-id" {
			t.Fatalf("second = %#v", second)
		}
		_, _ = io.WriteString(w, `{"ok":true,"result":[{"message_id":1,"chat":{"id":1,"type":"private"}},{"message_id":2,"chat":{"id":1,"type":"private"}}]}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	messages, err := client.SendMediaGroup(context.Background(), SendMediaGroupParams{
		ChatID: 1,
		Media: []MediaGroupItem{
			InputMediaPhoto{Media: "photo-id", Caption: "one"},
			InputMediaVideo{Media: "video-id", SupportsStreaming: true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].MessageID != 2 {
		t.Fatalf("messages = %#v", messages)
	}
}

func TestSendMediaGroupStreamsMultipleUploads(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reader, err := r.MultipartReader()
		if err != nil {
			t.Fatal(err)
		}
		fields := map[string]string{}
		files := map[string]string{}
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(part)
			if err != nil {
				t.Fatal(err)
			}
			if part.FileName() == "" {
				fields[part.FormName()] = string(data)
			} else {
				files[part.FormName()] = string(data)
			}
		}
		if files["a"] != "alpha" || files["b"] != "beta" {
			t.Fatalf("files = %#v", files)
		}
		var media []map[string]any
		if err := json.Unmarshal([]byte(fields["media"]), &media); err != nil {
			t.Fatal(err)
		}
		if media[0]["media"] != "attach://a" || media[1]["media"] != "attach://b" {
			t.Fatalf("media = %#v", media)
		}
		_, _ = io.WriteString(w, `{"ok":true,"result":[{"message_id":1,"chat":{"id":1,"type":"private"}},{"message_id":2,"chat":{"id":1,"type":"private"}}]}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	_, err := client.SendMediaGroupUpload(context.Background(), SendMediaGroupParams{
		ChatID: 1,
		Media: []MediaGroupItem{
			InputMediaPhoto{Media: Attachment("a")},
			InputMediaPhoto{Media: Attachment("b")},
		},
	},
		NewUpload("a", "a.jpg", strings.NewReader("alpha")),
		NewUpload("b", "b.jpg", strings.NewReader("beta")),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSendPollPreservesExplicitFalse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		value, exists := body["is_anonymous"]
		if !exists || value != false {
			t.Fatalf("is_anonymous = %#v, exists=%v", value, exists)
		}
		_, _ = io.WriteString(w, `{"ok":true,"result":{"message_id":1,"chat":{"id":1,"type":"private"},"poll":{"id":"p","question":"Q","options":[],"total_voter_count":0,"is_closed":false,"is_anonymous":false,"type":"regular","allows_multiple_answers":false,"allows_revoting":true,"members_only":false}}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	_, err := client.SendPoll(context.Background(), SendPollParams{
		ChatID: 1, Question: "Q",
		Options:     []InputPollOption{{Text: "A"}},
		IsAnonymous: Bool(false),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAdministrationAndReactionPayloads(t *testing.T) {
	t.Parallel()

	tests := []struct {
		method string
		call   func(context.Context, *Client) error
		check  func(*testing.T, map[string]any)
	}{
		{
			method: "restrictChatMember",
			call: func(ctx context.Context, client *Client) error {
				return client.RestrictChatMember(ctx, RestrictChatMemberParams{
					ChatID: 1, UserID: 9,
					Permissions: ChatPermissions{CanSendMessages: true, CanReactToMessages: true, CanEditTag: true},
				})
			},
			check: func(t *testing.T, body map[string]any) {
				permissions := body["permissions"].(map[string]any)
				if permissions["can_send_messages"] != true || permissions["can_react_to_messages"] != true || permissions["can_edit_tag"] != true {
					t.Fatalf("permissions = %#v", permissions)
				}
			},
		},
		{
			method: "setMessageReaction",
			call: func(ctx context.Context, client *Client) error {
				return client.SetMessageReaction(ctx, SetMessageReactionParams{
					ChatID: 1, MessageID: 2, Reaction: []ReactionType{{Type: ReactionEmoji, Emoji: "🔥"}},
				})
			},
			check: func(t *testing.T, body map[string]any) {
				reactions := body["reaction"].([]any)
				if reactions[0].(map[string]any)["emoji"] != "🔥" {
					t.Fatalf("reaction = %#v", reactions)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasSuffix(r.URL.Path, "/"+test.method) {
					t.Fatalf("path = %s", r.URL.Path)
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				test.check(t, body)
				_, _ = io.WriteString(w, `{"ok":true,"result":true}`)
			}))
			defer server.Close()
			client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
			if err := test.call(context.Background(), client); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSendPollUploadUsesTypedMedia(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reader, err := r.MultipartReader()
		if err != nil {
			t.Fatal(err)
		}
		fields := map[string]string{}
		files := map[string]string{}
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(part)
			if err != nil {
				t.Fatal(err)
			}
			if part.FileName() == "" {
				fields[part.FormName()] = string(data)
			} else {
				files[part.FormName()] = string(data)
			}
		}
		if files["option-photo"] != "image" {
			t.Fatalf("files = %#v", files)
		}
		if fields["is_anonymous"] != "false" {
			t.Fatalf("is_anonymous = %q", fields["is_anonymous"])
		}
		var options []map[string]any
		if err := json.Unmarshal([]byte(fields["options"]), &options); err != nil {
			t.Fatal(err)
		}
		media := options[0]["media"].(map[string]any)
		if media["type"] != "photo" || media["media"] != "attach://option-photo" {
			t.Fatalf("option media = %#v", media)
		}
		var descriptionMedia map[string]any
		if err := json.Unmarshal([]byte(fields["media"]), &descriptionMedia); err != nil {
			t.Fatal(err)
		}
		if descriptionMedia["type"] != "location" {
			t.Fatalf("description media = %#v", descriptionMedia)
		}
		_, _ = io.WriteString(w, `{"ok":true,"result":{"message_id":1,"chat":{"id":1,"type":"private"},"poll":{"id":"p","question":"Q","options":[],"total_voter_count":0,"is_closed":false,"is_anonymous":false,"type":"regular","allows_multiple_answers":false,"allows_revoting":true,"members_only":false}}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	_, err := client.SendPollUpload(context.Background(), SendPollParams{
		ChatID: 1, Question: "Q", IsAnonymous: Bool(false),
		Options: []InputPollOption{{
			Text:  "A",
			Media: InputMediaPhoto{Media: Attachment("option-photo")},
		}},
		Media: InputMediaLocation{Latitude: 12.5, Longitude: 77.6},
	}, NewUpload("option-photo", "option.jpg", strings.NewReader("image")))
	if err != nil {
		t.Fatal(err)
	}
}

func TestAttachmentParityRejectsMissingAndExtraUploads(t *testing.T) {
	t.Parallel()

	media := []MediaGroupItem{
		InputMediaPhoto{Media: Attachment("required")},
		InputMediaPhoto{Media: "existing-id"},
	}
	if err := validateAttachmentUploads(media, nil, "sendMediaGroup"); err == nil || !strings.Contains(err.Error(), "has no upload") {
		t.Fatalf("missing upload error = %v", err)
	}
	if err := validateAttachmentUploads(media, []Upload{
		NewUpload("required", "a.jpg", strings.NewReader("a")),
		NewUpload("extra", "b.jpg", strings.NewReader("b")),
	}, "sendMediaGroup"); err == nil || !strings.Contains(err.Error(), "not referenced") {
		t.Fatalf("extra upload error = %v", err)
	}
}

func TestMediaGroupSupportsLivePhotos(t *testing.T) {
	t.Parallel()

	item := InputMediaLivePhoto{Media: "video-file", Photo: "photo-file", Caption: "live"}
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["type"] != "live_photo" || payload["photo"] != "photo-file" {
		t.Fatalf("payload = %#v", payload)
	}

	params := SendMediaGroupParams{ChatID: 1, Media: []MediaGroupItem{
		item,
		InputMediaPhoto{Media: "photo-id"},
	}}
	if err := validateMediaGroup(params); err != nil {
		t.Fatal(err)
	}
	params.Media = []MediaGroupItem{
		InputMediaLivePhoto{Media: "video-file"},
		InputMediaPhoto{Media: "photo-id"},
	}
	if err := validateMediaGroup(params); err == nil {
		t.Fatal("expected missing static-photo validation error")
	}
}

func TestCreateChatInviteLinkResult(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"ok":true,"result":{"invite_link":"https://t.me/+x","creator":{"id":1,"is_bot":true,"first_name":"Bot"},"creates_join_request":true,"is_primary":false,"is_revoked":false}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	link, err := client.CreateChatInviteLink(context.Background(), CreateChatInviteLinkParams{ChatID: 1, CreatesJoinRequest: true})
	if err != nil {
		t.Fatal(err)
	}
	if link.InviteLink != "https://t.me/+x" || !link.CreatesJoinRequest {
		t.Fatalf("link = %#v", link)
	}
}
