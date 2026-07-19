package types

import (
	"encoding/json"
	"testing"
)

func TestInlineQueryResultInjectsDiscriminator(t *testing.T) {
	t.Parallel()
	value := InlineQueryResultArticle{
		ID:                  "result-1",
		Title:               "Hermes",
		InputMessageContent: InputTextMessageContent{MessageText: "Fast"},
	}
	var result InlineQueryResult = value
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["type"] != "article" || decoded["id"] != "result-1" {
		t.Fatalf("inline result = %s", data)
	}
	content := decoded["input_message_content"].(map[string]any)
	if content["message_text"] != "Fast" {
		t.Fatalf("input content = %#v", content)
	}
}

func TestPassportElementErrorInjectsSource(t *testing.T) {
	t.Parallel()
	var value PassportElementError = PassportElementErrorDataField{
		Type: "personal_details", FieldName: "first_name", DataHash: "hash", Message: "Invalid",
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["source"] != "data" || decoded["type"] != "personal_details" {
		t.Fatalf("passport error = %s", data)
	}
}

func TestMaybeInaccessibleMessageDecode(t *testing.T) {
	t.Parallel()
	var accessible MaybeInaccessibleMessage
	if err := json.Unmarshal([]byte(`{"message_id":7,"date":1,"chat":{"id":2,"type":"private"}}`), &accessible); err != nil {
		t.Fatal(err)
	}
	if message, ok := accessible.Accessible(); !ok || message.MessageID != 7 {
		t.Fatalf("accessible = %#v", accessible)
	}

	var inaccessible MaybeInaccessibleMessage
	if err := json.Unmarshal([]byte(`{"message_id":8,"date":0,"chat":{"id":2,"type":"private"}}`), &inaccessible); err != nil {
		t.Fatal(err)
	}
	if message, ok := inaccessible.Inaccessible(); !ok || message.MessageID != 8 {
		t.Fatalf("inaccessible = %#v", inaccessible)
	}
}
