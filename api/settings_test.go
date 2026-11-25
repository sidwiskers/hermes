package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func boolSetting(value bool) *bool       { return &value }
func stringSetting(value string) *string { return &value }
func intSetting(value int) *int          { return &value }

func TestSetMySettingsPreservesExplicitOptionalValues(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/botTOKEN/setMySettings" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		expected := map[string]any{
			"supports_inline_queries":           false,
			"inline_query_placeholder":          "",
			"needs_inline_query_location":       true,
			"inline_feedback_share_denominator": float64(0),
			"can_join_groups":                   false,
			"can_read_all_group_messages":       true,
			"can_connect_to_business":           false,
			"can_manage_bots":                   true,
			"supports_guest_queries":            false,
			"supports_join_request_queries":     true,
			"has_topics_enabled":                true,
			"allows_users_to_create_topics":     false,
			"has_topic_sidebar":                 true,
			"can_communicate_with_bots":         false,
			"callback_query_mode":               "inline",
		}
		if len(body) != len(expected) {
			t.Fatalf("got %d settings, want %d: %#v", len(body), len(expected), body)
		}
		for key, want := range expected {
			if got, ok := body[key]; !ok || got != want {
				t.Errorf("%s = %#v (present=%v), want %#v", key, got, ok, want)
			}
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	params := SetMySettingsParams{
		SupportsInlineQueries:          boolSetting(false),
		InlineQueryPlaceholder:         stringSetting(""),
		NeedsInlineQueryLocation:       boolSetting(true),
		InlineFeedbackShareDenominator: intSetting(0),
		CanJoinGroups:                  boolSetting(false),
		CanReadAllGroupMessages:        boolSetting(true),
		CanConnectToBusiness:           boolSetting(false),
		CanManageBots:                  boolSetting(true),
		SupportsGuestQueries:           boolSetting(false),
		SupportsJoinRequestQueries:     boolSetting(true),
		HasTopicsEnabled:               boolSetting(true),
		AllowsUsersToCreateTopics:      boolSetting(false),
		HasTopicSidebar:                boolSetting(true),
		CanCommunicateWithBots:         boolSetting(false),
		CallbackQueryMode:              stringSetting("inline"),
	}
	if err := client.SetMySettings(context.Background(), params); err != nil {
		t.Fatal(err)
	}
}

func TestSetMySettingsValidation(t *testing.T) {
	t.Parallel()

	client := New("TOKEN")
	for _, params := range []SetMySettingsParams{
		{InlineFeedbackShareDenominator: intSetting(2)},
		{CallbackQueryMode: stringSetting("legacy")},
	} {
		if err := client.SetMySettings(context.Background(), params); err == nil {
			t.Fatal("expected invalid settings to be rejected before transport")
		}
	}
}

func TestGetMeDecodesBotFullInfo(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/botTOKEN/getMe" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"id":9007199254740991,"is_bot":true,"first_name":"Hermes","username":"hermes_bot","active_usernames":["hermes_bot","hermes"],"can_join_groups":true,"can_read_all_group_messages":false,"supports_guest_queries":true,"supports_inline_queries":true,"inline_query_placeholder":"Search","needs_inline_query_location":false,"can_connect_to_business":true,"has_main_web_app":true,"has_topics_enabled":true,"allows_users_to_create_topics":false,"has_topic_sidebar":true,"can_manage_bots":false,"supports_join_request_queries":true}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	info, err := client.GetMe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.ID != 9007199254740991 || !info.IsBot || info.Username != "hermes_bot" ||
		len(info.ActiveUsernames) != 2 || info.InlineQueryPlaceholder != "Search" ||
		!info.HasTopicSidebar || !info.CanJoinGroups || info.CanReadAllGroupMessages {
		t.Fatalf("bot info = %#v", info)
	}
}
