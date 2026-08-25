package types

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBotAPI103ReplyMarkup(t *testing.T) {
	t.Parallel()

	markup := InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{{{
			Text:     "Unavailable",
			Disabled: &DisabledButton{},
		}}},
		ForceReply: true,
	}
	data, err := json.Marshal(markup)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"inline_keyboard":[[{"text":"Unavailable","disabled":{}}]],"force_reply":true}` {
		t.Fatalf("inline markup = %s", data)
	}

	reply, err := json.Marshal(ReplyKeyboardMarkup{Keyboard: [][]KeyboardButton{{{Text: "Reply"}}}, ForceReply: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(reply), `"force_reply":true`) {
		t.Fatalf("reply markup = %s", reply)
	}
}

func TestBotAPI103UpdateAndServiceObjects(t *testing.T) {
	t.Parallel()

	var update Update
	err := json.Unmarshal([]byte(`{
		"update_id":10,
		"stopped_message_generation":{
			"chat":{"id":7,"type":"private"},
			"message_thread_id":3,
			"draft_id":99
		}
	}`), &update)
	if err != nil {
		t.Fatal(err)
	}
	if update.Type() != UpdateStoppedMessageGeneration || update.StoppedMessageGeneration == nil || update.StoppedMessageGeneration.DraftID != 99 {
		t.Fatalf("update = %#v", update)
	}

	var message Message
	err = json.Unmarshal([]byte(`{
		"message_id":1,
		"date":1,
		"chat":{"id":7,"type":"supergroup"},
		"community_chat_joined":{"community":{"id":8,"name":"Hermes"}}
	}`), &message)
	if err != nil {
		t.Fatal(err)
	}
	if message.CommunityChatJoined == nil || message.CommunityChatJoined.Community.Name != "Hermes" {
		t.Fatalf("message = %#v", message)
	}
}

func TestBotAPI103GeneratedFields(t *testing.T) {
	t.Parallel()

	rights := ChatAdministratorRights{CanSendWelcomeMessages: true}
	data, err := json.Marshal(rights)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"can_send_welcome_messages":true`) {
		t.Fatalf("rights = %s", data)
	}

	var gift UniqueGiftInfo
	err = json.Unmarshal([]byte(`{
		"gift":{"base_name":"Gift","name":"gift-1","number":1,"model":{},"symbol":{},"backdrop":{}},
		"origin":"gifted",
		"text":"private note",
		"entities":[],
		"is_private":true
	}`), &gift)
	if err != nil {
		t.Fatal(err)
	}
	if gift.Text != "private note" || !gift.IsPrivate {
		t.Fatalf("gift = %#v", gift)
	}
}
