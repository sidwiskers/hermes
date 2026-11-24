package types

import (
	"encoding/json"
	"testing"
)

func TestTypedReactionAndChatMemberDecoding(t *testing.T) {
	t.Parallel()

	var update Update
	payload := []byte(`{"update_id":1,"message_reaction":{"chat":{"id":1,"type":"supergroup"},"message_id":2,"user":{"id":3,"is_bot":false,"first_name":"A"},"date":4,"old_reaction":[],"new_reaction":[{"type":"emoji","emoji":"🔥"}]}}`)
	if err := json.Unmarshal(payload, &update); err != nil {
		t.Fatal(err)
	}
	if update.MessageReaction == nil || len(update.MessageReaction.NewReaction) != 1 || update.MessageReaction.NewReaction[0].Emoji != "🔥" {
		t.Fatalf("reaction update = %#v", update.MessageReaction)
	}

	var member ChatMember
	if err := json.Unmarshal([]byte(`{"status":"administrator","user":{"id":3,"is_bot":false,"first_name":"A"},"can_manage_chat":true,"can_delete_messages":true}`), &member); err != nil {
		t.Fatal(err)
	}
	if !member.IsAdministrator() || !member.CanDeleteMessages {
		t.Fatalf("member = %#v", member)
	}
}

func TestLivePhotoAndEditTagDecoding(t *testing.T) {
	t.Parallel()

	var message Message
	if err := json.Unmarshal([]byte(`{"message_id":1,"chat":{"id":1,"type":"private"},"live_photo":{"file_id":"v","file_unique_id":"u","width":640,"height":480,"duration":2}}`), &message); err != nil {
		t.Fatal(err)
	}
	if message.LivePhoto == nil || message.LivePhoto.FileID != "v" {
		t.Fatalf("live photo = %#v", message.LivePhoto)
	}

	var permissions ChatPermissions
	if err := json.Unmarshal([]byte(`{"can_edit_tag":true,"can_react_to_messages":true}`), &permissions); err != nil {
		t.Fatal(err)
	}
	if !permissions.CanEditTag || !permissions.CanReactToMessages {
		t.Fatalf("permissions = %#v", permissions)
	}
}


func TestCallbackQueryInlineMessageIDDecoding(t *testing.T) {
	t.Parallel()

	var update Update
	payload := []byte(`{"update_id":9,"callback_query":{"id":"callback-1","from":{"id":7,"is_bot":false,"first_name":"Ada"},"inline_message_id":"inline-message-42","chat_instance":"chat-instance","data":"next"}}`)
	if err := json.Unmarshal(payload, &update); err != nil {
		t.Fatal(err)
	}
	if update.CallbackQuery == nil || update.CallbackQuery.Message != nil ||
		update.CallbackQuery.InlineMessageID != "inline-message-42" ||
		update.CallbackQuery.Data != "next" {
		t.Fatalf("callback query = %#v", update.CallbackQuery)
	}
}

func TestBotAPI105ExpandableQuotationAndInviterDecoding(t *testing.T) {
	t.Parallel()

	var block RichBlockExpandableBlockQuotation
	payload := []byte(`{"type":"expandable_blockquote","blocks":[{"type":"paragraph","text":"Quote"},{"type":"divider"}],"credit":"Author"}`)
	if err := json.Unmarshal(payload, &block); err != nil {
		t.Fatal(err)
	}
	if block.Type != "expandable_blockquote" || len(block.Blocks) != 2 ||
		block.Blocks[0].Type != "paragraph" || block.Blocks[1].Type != "divider" {
		t.Fatalf("expandable quotation = %#v", block)
	}

	var allowed WriteAccessAllowed
	if err := json.Unmarshal([]byte(`{"inviter":{"id":55,"is_bot":false,"first_name":"Owner"}}`), &allowed); err != nil {
		t.Fatal(err)
	}
	if allowed.Inviter == nil || allowed.Inviter.ID != 55 || allowed.Inviter.FirstName != "Owner" {
		t.Fatalf("write access inviter = %#v", allowed.Inviter)
	}
}
