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

func TestInputRichBlocksEncodeDiscriminators(t *testing.T) {
	t.Parallel()

	message := InputRichMessage{Blocks: []InputRichBlock{
		InputRichBlockSectionHeading{Text: "Title", Size: 1},
		InputRichBlockDetails{
			Summary: "More",
			Blocks: []InputRichBlock{
				InputRichBlockParagraph{Text: "Body"},
				InputRichBlockDivider{},
			},
		},
	}}
	data, err := json.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"blocks":[{"type":"heading","text":"Title","size":1},{"type":"details","summary":"More","blocks":[{"type":"paragraph","text":"Body"},{"type":"divider"}]}]}`
	if string(data) != want {
		t.Fatalf("rich message = %s, want %s", data, want)
	}
}

func TestSendRichMessageRequestAndResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/botTOKEN/sendRichMessage" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		var params struct {
			ChatID      int64            `json:"chat_id"`
			RichMessage InputRichMessage `json:"rich_message"`
		}
		if err := json.NewDecoder(request.Body).Decode(&params); err != nil {
			t.Fatal(err)
		}
		if params.ChatID != 7 || params.RichMessage.HTML != "<p>Hello</p>" {
			t.Fatalf("params = %#v", params)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"message_id":1,"date":1,"chat":{"id":7,"type":"private"},"rich_message":{"blocks":[{"type":"paragraph","text":"Hello"}]}}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	message, err := client.SendRichMessage(context.Background(), SendRichMessageParams{
		ChatID:      int64(7),
		RichMessage: InputRichMessage{HTML: "<p>Hello</p>"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if message.RichMessage == nil || len(message.RichMessage.Blocks) != 1 || message.RichMessage.Blocks[0].Type != "paragraph" {
		t.Fatalf("message = %#v", message)
	}
}

func TestSendRichMessageStreamsNestedAttachment(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		rich := request.FormValue("rich_message")
		if !strings.Contains(rich, `"type":"photo"`) || !strings.Contains(rich, `"media":"attach://photo"`) {
			t.Fatalf("rich_message = %s", rich)
		}
		file, header, err := request.FormFile("photo")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		data, _ := io.ReadAll(file)
		if header.Filename != "photo.jpg" || string(data) != "image" {
			t.Fatalf("upload = %q %q", header.Filename, data)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"message_id":1,"date":1,"chat":{"id":7,"type":"private"}}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	_, err := client.SendRichMessageUpload(context.Background(), SendRichMessageParams{
		ChatID: int64(7),
		RichMessage: InputRichMessage{Blocks: []InputRichBlock{
			InputRichBlockPhoto{Photo: InputMediaPhoto{Media: Attachment("photo")}},
		}},
	}, NewUpload("photo", "photo.jpg", strings.NewReader("image")))
	if err != nil {
		t.Fatal(err)
	}
}

func TestRichMessageValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message InputRichMessage
		draft   bool
	}{
		{name: "missing format"},
		{name: "multiple formats", message: InputRichMessage{HTML: "x", Markdown: "x"}},
		{
			name: "media with blocks",
			message: InputRichMessage{
				Blocks: []InputRichBlock{InputRichBlockDivider{}},
				Media:  []InputRichMessageMedia{{ID: "x", Media: InputMediaPhoto{Media: "id"}}},
			},
		},
		{
			name: "invalid media id",
			message: InputRichMessage{
				HTML:  "x",
				Media: []InputRichMessageMedia{{ID: "bad id", Media: InputMediaPhoto{Media: "id"}}},
			},
		},
		{
			name:    "thinking final",
			message: InputRichMessage{Blocks: []InputRichBlock{InputRichBlockThinking{Text: "Thinking"}}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := validateRichMessage(test.message, test.draft); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
	if err := validateRichMessage(InputRichMessage{Blocks: []InputRichBlock{
		InputRichBlockThinking{Text: "Thinking"},
	}}, true); err != nil {
		t.Fatalf("draft thinking block: %v", err)
	}
}
