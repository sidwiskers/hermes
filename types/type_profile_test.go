package types

import (
	"encoding/json"
	"testing"
)

func TestTypedChatBoostUpdate(t *testing.T) {
	t.Parallel()

	var update Update
	err := json.Unmarshal([]byte(`{"update_id":1,"chat_boost":{"chat":{"id":7,"type":"supergroup"},"boost":{"boost_id":"boost-1","add_date":1,"expiration_date":2,"source":{"source":"premium","user":{"id":9,"is_bot":false,"first_name":"Ada"}}}}}`), &update)
	if err != nil {
		t.Fatal(err)
	}
	if update.Type() != UpdateChatBoost || update.ChatBoost == nil || update.ChatBoost.Boost.Source.User == nil {
		t.Fatalf("update = %#v", update)
	}
}

func TestTypedManagedBotUpdate(t *testing.T) {
	t.Parallel()

	var update Update
	err := json.Unmarshal([]byte(`{"update_id":2,"managed_bot":{"user":{"id":1,"is_bot":false,"first_name":"Owner"},"bot":{"id":2,"is_bot":true,"first_name":"Worker"}}}`), &update)
	if err != nil {
		t.Fatal(err)
	}
	if update.Type() != UpdateManagedBot || update.Sender() == nil || update.Sender().ID != 1 {
		t.Fatalf("update = %#v", update)
	}
}
