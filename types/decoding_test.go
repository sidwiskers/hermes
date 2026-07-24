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
