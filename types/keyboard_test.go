package types

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReplyMarkupVariantsMarshal(t *testing.T) {
	t.Parallel()

	markup := ReplyKeyboard(KeyRow(Key("A"), Key("B")))
	data, err := json.Marshal(markup)
	if err != nil {
		t.Fatal(err)
	}
	payload := string(data)
	if !strings.Contains(payload, `"keyboard"`) || !strings.Contains(payload, `"resize_keyboard":true`) {
		t.Fatalf("payload = %s", payload)
	}
}
