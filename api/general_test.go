package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestForwardMessagesContract(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/botTOKEN/forwardMessages" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		var params ForwardMessagesParams
		if err := json.NewDecoder(request.Body).Decode(&params); err != nil {
			t.Fatal(err)
		}
		if len(params.MessageIDs) != 3 || params.MessageIDs[2] != 9 || params.DirectMessagesTopicID != 4 {
			t.Fatalf("params = %#v", params)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":[{"message_id":20},{"message_id":21}]}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	messageIDs, err := client.ForwardMessages(context.Background(), ForwardMessagesParams{
		ChatID: 7, FromChatID: 8, MessageIDs: []int{1, 5, 9}, DirectMessagesTopicID: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(messageIDs) != 2 || messageIDs[1].MessageID != 21 {
		t.Fatalf("message IDs = %#v", messageIDs)
	}
}

func TestBulkMessageValidation(t *testing.T) {
	t.Parallel()

	client := New("TOKEN")
	_, err := client.CopyMessages(context.Background(), CopyMessagesParams{
		ChatID: 1, FromChatID: 2, MessageIDs: []int{3, 3},
	})
	if err == nil {
		t.Fatal("expected strictly increasing validation error")
	}
	if err := client.SendMessageDraft(context.Background(), SendMessageDraftParams{ChatID: 1}); err == nil {
		t.Fatal("expected missing draft_id error")
	}
}

func TestLogOutUsesParameterlessMethod(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/botTOKEN/logOut" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	if err := client.LogOut(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestSetMySettingsEncodesEveryOptionalSetting(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/botTOKEN/setMySettings" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		var body map[string]json.RawMessage
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		want := map[string]string{
			"supports_inline_queries": "true",
			"inline_query_placeholder": `""`,
			"needs_inline_query_location": "false",
			"inline_feedback_share_denominator": "0",
			"can_join_groups": "false",
			"can_read_all_group_messages": "true",
			"can_connect_to_business": "false",
			"can_manage_bots": "true",
			"supports_guest_queries": "true",
			"supports_join_request_queries": "false",
			"has_topics_enabled": "true",
			"allows_users_to_create_topics": "false",
			"has_topic_sidebar": "true",
			"can_communicate_with_bots": "true",
			"callback_query_mode": `"inline"`,
		}
		if len(body) != len(want) {
			t.Fatalf("encoded %d settings, want %d: %#v", len(body), len(want), body)
		}
		for key, expected := range want {
			if actual := string(body[key]); actual != expected {
				t.Errorf("%s = %s, want %s", key, actual, expected)
			}
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	placeholder := ""
	zero := 0
	if err := client.SetMySettings(context.Background(), SetMySettingsParams{
		SupportsInlineQueries:         Bool(true),
		InlineQueryPlaceholder:         &placeholder,
		NeedsInlineQueryLocation:       Bool(false),
		InlineFeedbackShareDenominator: &zero,
		CanJoinGroups:                  Bool(false),
		CanReadAllGroupMessages:        Bool(true),
		CanConnectToBusiness:           Bool(false),
		CanManageBots:                  Bool(true),
		SupportsGuestQueries:           Bool(true),
		SupportsJoinRequestQueries:     Bool(false),
		HasTopicsEnabled:               Bool(true),
		AllowsUsersToCreateTopics:      Bool(false),
		HasTopicSidebar:                Bool(true),
		CanCommunicateWithBots:         Bool(true),
		CallbackQueryMode:              CallbackQueryModeInline,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSetMySettingsRejectsInvalidEnumeratedValues(t *testing.T) {
	t.Parallel()

	client := New("TOKEN")
	invalidDenominator := 7
	if err := client.SetMySettings(context.Background(), SetMySettingsParams{
		InlineFeedbackShareDenominator: &invalidDenominator,
	}); err == nil {
		t.Fatal("expected invalid inline_feedback_share_denominator to fail")
	}
	if err := client.SetMySettings(context.Background(), SetMySettingsParams{
		CallbackQueryMode: CallbackQueryMode("sometimes"),
	}); err == nil {
		t.Fatal("expected invalid callback_query_mode to fail")
	}
}

func TestGetMeDecodesBotFullInfo(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/botTOKEN/getMe" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{
			"id":42,"is_bot":true,"first_name":"Hermes","username":"hermes_bot",
			"active_usernames":["hermes_bot","hermes_agent"],
			"can_join_groups":true,"can_read_all_group_messages":false,
			"supports_guest_queries":true,"supports_inline_queries":true,
			"inline_query_placeholder":"Search","needs_inline_query_location":false,
			"can_connect_to_business":true,"has_main_web_app":false,
			"has_topics_enabled":true,"allows_users_to_create_topics":true,
			"has_topic_sidebar":true,"can_manage_bots":false,
			"supports_join_request_queries":true
		}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	me, err := client.GetMe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if me.ID != 42 || me.Username != "hermes_bot" || !me.IsBot {
		t.Fatalf("identity = %#v", me)
	}
	if len(me.ActiveUsernames) != 2 || me.ActiveUsernames[1] != "hermes_agent" {
		t.Fatalf("active usernames = %#v", me.ActiveUsernames)
	}
	if me.InlineQueryPlaceholder != "Search" || !me.HasTopicSidebar || !me.HasTopicsEnabled {
		t.Fatalf("Bot API 10.5 settings missing from getMe: %#v", me)
	}
}
