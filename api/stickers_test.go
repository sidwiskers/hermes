package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateNewStickerSetStreamsNestedStickers(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		stickers := request.FormValue("stickers")
		if !strings.Contains(stickers, `"format":"static"`) || !strings.Contains(stickers, "attach://first") {
			t.Fatalf("stickers = %s", stickers)
		}
		file, _, err := request.FormFile("first")
		if err != nil {
			t.Fatal(err)
		}
		_ = file.Close()
		_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	err := client.CreateNewStickerSetUpload(context.Background(), CreateNewStickerSetParams{
		UserID: 9, Name: "hermes_by_bot", Title: "Hermes",
		Stickers: []InputSticker{{Sticker: Attachment("first"), Format: StickerFormatStatic, EmojiList: []string{"⚡"}}},
	}, NewUpload("first", "first.webp", strings.NewReader("image")))
	if err != nil {
		t.Fatal(err)
	}
}

func TestStickerValidation(t *testing.T) {
	t.Parallel()

	client := New("TOKEN")
	err := client.CreateNewStickerSet(context.Background(), CreateNewStickerSetParams{
		UserID: 1, Name: "set", Title: "Set",
		Stickers: []InputSticker{{Sticker: "file", Format: "bad", EmojiList: []string{"⚡"}}},
	})
	if err == nil {
		t.Fatal("expected invalid sticker format")
	}
	if _, err := client.GetCustomEmojiStickers(context.Background(), GetCustomEmojiStickersParams{}); err == nil {
		t.Fatal("expected empty identifier validation")
	}
}
