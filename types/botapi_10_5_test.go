package types

import (
	"encoding/json"
	"testing"
)

func TestBotAPI105BotFullInfoDecode(t *testing.T) {
	t.Parallel()

	var info BotFullInfo
	payload := []byte(`{
		"id":42,
		"is_bot":true,
		"first_name":"Hermes",
		"username":"hermes_bot",
		"active_usernames":["hermes_bot","hermes_agent"],
		"can_join_groups":true,
		"can_read_all_group_messages":false,
		"supports_guest_queries":true,
		"supports_inline_queries":true,
		"inline_query_placeholder":"Search inline",
		"needs_inline_query_location":false,
		"can_connect_to_business":true,
		"has_main_web_app":false,
		"has_topics_enabled":true,
		"allows_users_to_create_topics":false,
		"has_topic_sidebar":true,
		"can_manage_bots":true,
		"supports_join_request_queries":false
	}`)
	if err := json.Unmarshal(payload, &info); err != nil {
		t.Fatal(err)
	}
	var botID int64 = info.ID
	if info.ID != 42 || botID != 42 || !info.IsBot || info.Username != "hermes_bot" {
		t.Fatalf("identity fields = %#v", info)
	}
	if len(info.ActiveUsernames) != 2 || info.ActiveUsernames[1] != "hermes_agent" {
		t.Fatalf("active_usernames = %#v", info.ActiveUsernames)
	}
	if info.InlineQueryPlaceholder != "Search inline" || !info.HasTopicSidebar || !info.HasTopicsEnabled {
		t.Fatalf("Bot API 10.5 fields = %#v", info)
	}
}

func TestBotAPI105InviterOnWriteAccessAllowed(t *testing.T) {
	t.Parallel()

	var update Update
	payload := []byte(`{
		"update_id":105,
		"message":{
			"message_id":9,
			"date":1,
			"chat":{"id":100,"type":"private"},
			"write_access_allowed":{
				"inviter":{"id":7,"is_bot":false,"first_name":"Owner"}
			}
		}
	}`)
	if err := json.Unmarshal(payload, &update); err != nil {
		t.Fatal(err)
	}
	if update.Message == nil || update.Message.WriteAccessAllowed == nil {
		t.Fatalf("service message = %#v", update.Message)
	}
	inviter := update.Message.WriteAccessAllowed.Inviter
	if inviter == nil || inviter.ID != 7 || inviter.FirstName != "Owner" {
		t.Fatalf("inviter = %#v", inviter)
	}
}

func TestBotAPI105ExpandableBlockQuotationDecode(t *testing.T) {
	t.Parallel()

	var message Message
	payload := []byte(`{
		"message_id":5,
		"date":1,
		"chat":{"id":1,"type":"private"},
		"rich_message":{"blocks":[{
			"type":"expandable_blockquote",
			"credit":"Hermes",
			"blocks":[
				{"type":"paragraph","text":"Intro"},
				{"type":"table","cells":[[{"text":"A"},{"text":"B"}]]},
				{"type":"slideshow","blocks":[{"type":"paragraph","text":"Nested"}]}
			]
		}]}
	}`)
	if err := json.Unmarshal(payload, &message); err != nil {
		t.Fatal(err)
	}
	if message.RichMessage == nil || len(message.RichMessage.Blocks) != 1 {
		t.Fatalf("rich message = %#v", message.RichMessage)
	}
	quote := message.RichMessage.Blocks[0]
	if quote.Type != "expandable_blockquote" || quote.Credit != "Hermes" || len(quote.Blocks) != 3 {
		t.Fatalf("expandable quotation = %#v", quote)
	}
	if quote.Blocks[1].Type != "table" || quote.Blocks[2].Type != "slideshow" ||
		len(quote.Blocks[2].Blocks) != 1 || quote.Blocks[2].Blocks[0].Text != "Nested" {
		t.Fatalf("nested blocks = %#v", quote.Blocks)
	}
}
