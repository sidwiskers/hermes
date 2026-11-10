package types

import (
	"encoding/json"
	"testing"
)

func TestBotAPI104WidgetUnionsDecodeAndRoundTrip(t *testing.T) {
	t.Parallel()

	var query WidgetQuery
	payload := `{"id":"query-1","from":{"id":42,"is_bot":false,"first_name":"Tester"},"chat_instance":"chat-instance","type":{"type":"sign_data","wallet":{"address":"EQ...","amount":"10","public_key":"key"},"sign_data":"signed-payload"},"data":"widget-data","widget_id":"widget-1"}`
	if err := json.Unmarshal([]byte(payload), &query); err != nil {
		t.Fatal(err)
	}
	if query.Type.Kind() != "sign_data" || query.Type.SignData == nil {
		t.Fatalf("widget query type = %#v", query.Type)
	}
	if query.Type.SignData.Wallet.Address != "EQ..." || query.Type.SignData.SignData != "signed-payload" {
		t.Fatalf("sign data = %#v", query.Type.SignData)
	}
	encoded, err := json.Marshal(query)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip WidgetQuery
	if err := json.Unmarshal(encoded, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if roundTrip.Type.Kind() != "sign_data" || roundTrip.Type.SignData.Wallet.PublicKey != "key" {
		t.Fatalf("round-trip query type = %#v", roundTrip.Type)
	}
}

func TestBotAPI104ChatShareThemeAndEmojiUnions(t *testing.T) {
	t.Parallel()

	var theme ChatShareTheme
	payload := `{"type":"emoji","emojis":[{"type":"emoji","emoji":"🔥"},{"type":"custom_emoji","custom_emoji_id":"custom-1"}],"background_backdrop_id":9}`
	if err := json.Unmarshal([]byte(payload), &theme); err != nil {
		t.Fatal(err)
	}
	if theme.Kind() != "emoji" || theme.Emoji == nil || len(theme.Emoji.Emojis) != 2 {
		t.Fatalf("theme = %#v", theme)
	}
	if theme.Emoji.Emojis[0].Emoji == nil || theme.Emoji.Emojis[0].Emoji.Emoji != "🔥" {
		t.Fatalf("standard emoji = %#v", theme.Emoji.Emojis[0])
	}
	if theme.Emoji.Emojis[1].CustomEmoji == nil || theme.Emoji.Emojis[1].CustomEmoji.CustomEmojiID != "custom-1" {
		t.Fatalf("custom emoji = %#v", theme.Emoji.Emojis[1])
	}

	encoded, err := json.Marshal(theme)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["type"] != "emoji" || decoded["background_backdrop_id"] != float64(9) {
		t.Fatalf("encoded theme = %s", encoded)
	}
}

func TestBotAPI104UnknownWidgetUnionVariantIsPreserved(t *testing.T) {
	t.Parallel()

	const payload = `{"type":"future_variant","future_value":{"enabled":true}}`
	var value WidgetQueryType
	if err := json.Unmarshal([]byte(payload), &value); err != nil {
		t.Fatal(err)
	}
	if value.Kind() != "" || len(value.Raw) == 0 {
		t.Fatalf("unknown variant not retained: %#v", value)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var before, after any
	if err := json.Unmarshal([]byte(payload), &before); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &after); err != nil {
		t.Fatal(err)
	}
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)
	if string(beforeJSON) != string(afterJSON) {
		t.Fatalf("unknown variant changed: before=%s after=%s", beforeJSON, afterJSON)
	}
}

func TestBotAPI104RichTextTonAddressDiscriminator(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(RichTextTonAddress{
		Text:    RichTextBold{Text: "Wallet"},
		Address: "EQ...",
	})
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["type"] != "ton_address" || decoded["address"] != "EQ..." {
		t.Fatalf("TON rich text = %s", data)
	}
}

func TestBotAPI104WidgetMessageContentImplementsInlineUnion(t *testing.T) {
	t.Parallel()

	var content InputMessageContent = InputWidgetMessageContent{HTML: "<main>widget</main>"}
	data, err := json.Marshal(content)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["html"] != "<main>widget</main>" {
		t.Fatalf("widget content = %s", data)
	}

	var shared InputMessageContent = InputSharedWidgetMessageContent{WidgetID: "shared-1", ChatData: "state"}
	data, err = json.Marshal(shared)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["widget_id"] != "shared-1" || decoded["chat_data"] != "state" {
		t.Fatalf("shared widget content = %s", data)
	}
}

func TestBotAPI104WidgetUpdateIsClassifiedAndAttributed(t *testing.T) {
	t.Parallel()

	var update Update
	payload := `{"update_id":88,"widget_query":{"id":"query-2","from":{"id":77,"is_bot":false,"first_name":"Widget User"},"chat_instance":"instance","type":{"type":"data"},"data":"state"}}`
	if err := json.Unmarshal([]byte(payload), &update); err != nil {
		t.Fatal(err)
	}
	if update.Type() != UpdateWidgetQuery {
		t.Fatalf("update type = %q", update.Type())
	}
	sender := update.Sender()
	if sender == nil || sender.ID != 77 {
		t.Fatalf("update sender = %#v", sender)
	}
}
