package types

import (
	"encoding/json"
	"testing"
)

func TestRichTextDiscriminators(t *testing.T) {
	t.Parallel()

	value := []RichText{
		"plain ",
		RichTextBold{Text: "bold"},
		RichTextURL{Text: RichTextItalic{Text: "link"}, URL: "https://example.com"},
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	want := `["plain ",{"type":"bold","text":"bold"},{"type":"url","text":{"type":"italic","text":"link"},"url":"https://example.com"}]`
	if string(data) != want {
		t.Fatalf("rich text = %s, want %s", data, want)
	}
}

func TestRichMessageResponseDecode(t *testing.T) {
	t.Parallel()

	var message RichMessage
	err := json.Unmarshal([]byte(`{"blocks":[{"type":"paragraph","text":"hello"},{"type":"photo","photo":[{"file_id":"f","file_unique_id":"u","width":1,"height":1}]}],"is_rtl":true}`), &message)
	if err != nil {
		t.Fatal(err)
	}
	if !message.IsRTL || len(message.Blocks) != 2 || message.Blocks[0].Type != "paragraph" || message.Blocks[1].Photo[0].FileID != "f" {
		t.Fatalf("unexpected rich message: %#v", message)
	}
}
