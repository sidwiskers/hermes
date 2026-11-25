package api

import (
	"context"
	"fmt"
)

// SetMySettingsParams contains optional updates to a bot's global settings.
// Pointers distinguish omitted settings from explicit false, zero, or empty values.
type SetMySettingsParams struct {
	SupportsInlineQueries          *bool   `json:"supports_inline_queries,omitempty"`
	InlineQueryPlaceholder         *string `json:"inline_query_placeholder,omitempty"`
	NeedsInlineQueryLocation       *bool   `json:"needs_inline_query_location,omitempty"`
	InlineFeedbackShareDenominator *int    `json:"inline_feedback_share_denominator,omitempty"`
	CanJoinGroups                  *bool   `json:"can_join_groups,omitempty"`
	CanReadAllGroupMessages        *bool   `json:"can_read_all_group_messages,omitempty"`
	CanConnectToBusiness           *bool   `json:"can_connect_to_business,omitempty"`
	CanManageBots                  *bool   `json:"can_manage_bots,omitempty"`
	SupportsGuestQueries           *bool   `json:"supports_guest_queries,omitempty"`
	SupportsJoinRequestQueries     *bool   `json:"supports_join_request_queries,omitempty"`
	HasTopicsEnabled               *bool   `json:"has_topics_enabled,omitempty"`
	AllowsUsersToCreateTopics      *bool   `json:"allows_users_to_create_topics,omitempty"`
	HasTopicSidebar                *bool   `json:"has_topic_sidebar,omitempty"`
	CanCommunicateWithBots         *bool   `json:"can_communicate_with_bots,omitempty"`
	CallbackQueryMode              *string `json:"callback_query_mode,omitempty"`
}

// SetMySettings changes the bot's global settings. Only non-nil fields are sent;
// nil leaves the existing setting unchanged.
func (client *Client) SetMySettings(ctx context.Context, params SetMySettingsParams) error {
	if params.InlineFeedbackShareDenominator != nil {
		switch *params.InlineFeedbackShareDenominator {
		case 0, 1, 10, 100, 1000:
		default:
			return fmt.Errorf("hermes: setMySettings inline_feedback_share_denominator must be 0, 1, 10, 100, or 1000")
		}
	}
	if params.CallbackQueryMode != nil {
		switch *params.CallbackQueryMode {
		case "default", "auto", "inline":
		default:
			return fmt.Errorf("hermes: setMySettings callback_query_mode must be default, auto, or inline")
		}
	}
	return client.callTrue(ctx, "setMySettings", params)
}
