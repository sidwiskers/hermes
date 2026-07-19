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

func TestPostStoryStreamsTypedContent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		content := request.FormValue("content")
		if !strings.Contains(content, `"type":"photo"`) || !strings.Contains(content, "attach://story") {
			t.Fatalf("content = %s", content)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"chat":{"id":7,"type":"private"},"id":3}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	story, err := client.PostStory(context.Background(), PostStoryParams{
		BusinessConnectionID: "business", Content: InputStoryContentPhoto{Photo: Attachment("story")}, ActivePeriod: StoryActive24Hours,
	}, NewUpload("story", "story.jpg", strings.NewReader("image")))
	if err != nil {
		t.Fatal(err)
	}
	if story.ID != 3 {
		t.Fatalf("story = %#v", story)
	}
}

func TestSendChecklistContract(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		checklist := body["checklist"].(map[string]any)
		tasks := checklist["tasks"].([]any)
		if checklist["title"] != "Release" || len(tasks) != 2 {
			t.Fatalf("checklist = %#v", checklist)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"message_id":1,"chat":{"id":7,"type":"private"},"checklist":{"title":"Release","tasks":[]}}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	message, err := client.SendChecklist(context.Background(), SendChecklistParams{
		BusinessConnectionID: "business", ChatID: 7,
		Checklist: InputChecklist{Title: "Release", Tasks: []InputChecklistTask{{ID: 1, Text: "Tag"}, {ID: 2, Text: "Publish"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if message.Checklist == nil || message.Checklist.Title != "Release" {
		t.Fatalf("message = %#v", message)
	}
}

func TestGetUserGiftsFlattensFilter(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["exclude_unique"] != true || body["limit"] != float64(10) {
			t.Fatalf("body = %#v", body)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"total_count":0,"gifts":[]}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	_, err := client.GetUserGifts(context.Background(), GetUserGiftsParams{
		UserID: 7, OwnedGiftsFilter: OwnedGiftsFilter{ExcludeUnique: true, Limit: 10},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetMyProfilePhotoInjectsType(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		photo := request.FormValue("photo")
		if !strings.Contains(photo, `"type":"static"`) || !strings.Contains(photo, "attach://avatar") {
			t.Fatalf("photo = %s", photo)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	err := client.SetMyProfilePhoto(context.Background(), SetMyProfilePhotoParams{
		Photo: InputProfilePhotoStatic{Photo: Attachment("avatar")},
	}, NewUpload("avatar", "avatar.jpg", strings.NewReader("image")))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSendPhotoUploadCarriesDirectMessageFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if request.FormValue("direct_messages_topic_id") != "12" {
			t.Fatalf("direct_messages_topic_id = %q", request.FormValue("direct_messages_topic_id"))
		}
		if !strings.Contains(request.FormValue("suggested_post_parameters"), `"price"`) {
			t.Fatalf("suggested_post_parameters = %q", request.FormValue("suggested_post_parameters"))
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"message_id":1,"chat":{"id":7,"type":"private"}}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	_, err := client.SendPhotoUpload(context.Background(), SendPhotoParams{
		ChatID: 7, DirectMessagesTopicID: 12,
		SuggestedPostParameters: &SuggestedPostParameters{Price: &SuggestedPostPrice{Currency: "XTR", Amount: 25}},
	}, "photo.jpg", strings.NewReader("image"))
	if err != nil {
		t.Fatal(err)
	}
}
