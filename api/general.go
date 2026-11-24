package api

import (
	"context"
	"fmt"
	"unicode/utf8"
)

// LogOut logs the bot out of Telegram's cloud Bot API server before it is
// launched against a local Bot API server.
func (client *Client) LogOut(ctx context.Context) error {
	return client.callTrue(ctx, "logOut", nil)
}

// Close closes a bot instance before it is moved between local Bot API
// servers. This is the Telegram Bot API close method; it does not close the
// underlying HTTP client.
func (client *Client) Close(ctx context.Context) error {
	return client.callTrue(ctx, "close", nil)
}

// CallbackQueryMode controls how callback queries from the bot's own messages
// identify their originating message.
type CallbackQueryMode = string

const (
	// CallbackQueryModeDefault always includes Message, using an inaccessible
	// message when Telegram cannot provide the full message.
	CallbackQueryModeDefault CallbackQueryMode = "default"
	// CallbackQueryModeAuto prefers InlineMessageID when the originating
	// message is inaccessible, but may still include an inaccessible Message
	// immediately after deletion.
	CallbackQueryModeAuto CallbackQueryMode = "auto"
	// CallbackQueryModeInline always uses InlineMessageID instead of Message
	// for callback queries from messages sent directly by the bot.
	CallbackQueryModeInline CallbackQueryMode = "inline"
)

// SetMySettingsParams contains the optional bot settings that can be changed
// with setMySettings. Pointer fields preserve explicit false and zero values;
// nil leaves the corresponding setting unchanged.
type SetMySettingsParams struct {
	SupportsInlineQueries          *bool              `json:"supports_inline_queries,omitempty"`
	InlineQueryPlaceholder          *string            `json:"inline_query_placeholder,omitempty"`
	NeedsInlineQueryLocation        *bool              `json:"needs_inline_query_location,omitempty"`
	InlineFeedbackShareDenominator  *int               `json:"inline_feedback_share_denominator,omitempty"`
	CanJoinGroups                   *bool              `json:"can_join_groups,omitempty"`
	CanReadAllGroupMessages         *bool              `json:"can_read_all_group_messages,omitempty"`
	CanConnectToBusiness            *bool              `json:"can_connect_to_business,omitempty"`
	CanManageBots                   *bool              `json:"can_manage_bots,omitempty"`
	SupportsGuestQueries            *bool              `json:"supports_guest_queries,omitempty"`
	SupportsJoinRequestQueries      *bool              `json:"supports_join_request_queries,omitempty"`
	HasTopicsEnabled                *bool              `json:"has_topics_enabled,omitempty"`
	AllowsUsersToCreateTopics       *bool              `json:"allows_users_to_create_topics,omitempty"`
	HasTopicSidebar                 *bool              `json:"has_topic_sidebar,omitempty"`
	CanCommunicateWithBots          *bool              `json:"can_communicate_with_bots,omitempty"`
	CallbackQueryMode               string             `json:"callback_query_mode,omitempty"`
}

// SetMySettings updates the bot's settings. Unspecified fields remain
// unchanged. Use Bool(false) to explicitly disable a boolean setting and a
// non-nil pointer to zero to disable inline feedback.
func (client *Client) SetMySettings(ctx context.Context, params SetMySettingsParams) error {
	if params.InlineFeedbackShareDenominator != nil {
		switch *params.InlineFeedbackShareDenominator {
		case 0, 1, 10, 100, 1000:
		default:
			return fmt.Errorf("hermes: setMySettings inline_feedback_share_denominator must be 0, 1, 10, 100, or 1000")
		}
	}
	switch params.CallbackQueryMode {
	case "", CallbackQueryModeDefault, CallbackQueryModeAuto, CallbackQueryModeInline:
	default:
		return fmt.Errorf("hermes: setMySettings callback_query_mode must be default, auto, or inline")
	}
	return client.callTrue(ctx, "setMySettings", params)
}

type ForwardMessagesParams struct {
	ChatID                any   `json:"chat_id"`
	MessageThreadID       int   `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID int   `json:"direct_messages_topic_id,omitempty"`
	FromChatID            any   `json:"from_chat_id"`
	MessageIDs            []int `json:"message_ids"`
	DisableNotification   bool  `json:"disable_notification,omitempty"`
	ProtectContent        bool  `json:"protect_content,omitempty"`
}

func validateBulkMessageIDs(messageIDs []int, method string) error {
	if len(messageIDs) == 0 || len(messageIDs) > 100 {
		return fmt.Errorf("hermes: %s requires 1-100 message_ids", method)
	}
	previous := 0
	for index, messageID := range messageIDs {
		if messageID <= 0 {
			return fmt.Errorf("hermes: %s message_ids must be positive", method)
		}
		if index != 0 && messageID <= previous {
			return fmt.Errorf("hermes: %s message_ids must be strictly increasing", method)
		}
		previous = messageID
	}
	return nil
}

func validateBulkMessagesTarget(chatID, fromChatID any, messageIDs []int, method string) error {
	if err := validateChatID(chatID, method); err != nil {
		return err
	}
	if err := validateChatID(fromChatID, method); err != nil {
		return fmt.Errorf("hermes: %s from_chat_id is required", method)
	}
	return validateBulkMessageIDs(messageIDs, method)
}

func (client *Client) ForwardMessages(ctx context.Context, params ForwardMessagesParams) ([]MessageID, error) {
	if err := validateBulkMessagesTarget(params.ChatID, params.FromChatID, params.MessageIDs, "forwardMessages"); err != nil {
		return nil, err
	}
	return Call[[]MessageID](ctx, client, "forwardMessages", params)
}

type CopyMessagesParams struct {
	ChatID                any   `json:"chat_id"`
	MessageThreadID       int   `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID int   `json:"direct_messages_topic_id,omitempty"`
	FromChatID            any   `json:"from_chat_id"`
	MessageIDs            []int `json:"message_ids"`
	DisableNotification   bool  `json:"disable_notification,omitempty"`
	ProtectContent        bool  `json:"protect_content,omitempty"`
	RemoveCaption         bool  `json:"remove_caption,omitempty"`
}

func (client *Client) CopyMessages(ctx context.Context, params CopyMessagesParams) ([]MessageID, error) {
	if err := validateBulkMessagesTarget(params.ChatID, params.FromChatID, params.MessageIDs, "copyMessages"); err != nil {
		return nil, err
	}
	return Call[[]MessageID](ctx, client, "copyMessages", params)
}

type SendMessageDraftParams struct {
	ChatID          int64           `json:"chat_id"`
	MessageThreadID int             `json:"message_thread_id,omitempty"`
	DraftID         int64           `json:"draft_id"`
	Text            string          `json:"text,omitempty"`
	ParseMode       string          `json:"parse_mode,omitempty"`
	Entities        []MessageEntity `json:"entities,omitempty"`
	CanStop         bool            `json:"can_stop,omitempty"`
	KeepOnStop      bool            `json:"keep_on_stop,omitempty"`
}

func (client *Client) SendMessageDraft(ctx context.Context, params SendMessageDraftParams) error {
	if params.ChatID == 0 {
		return fmt.Errorf("hermes: sendMessageDraft chat_id is required")
	}
	if params.DraftID == 0 {
		return fmt.Errorf("hermes: sendMessageDraft draft_id must be non-zero")
	}
	if utf8.RuneCountInString(params.Text) > 4096 {
		return fmt.Errorf("hermes: sendMessageDraft text must not exceed 4096 characters")
	}
	return client.callTrue(ctx, "sendMessageDraft", params)
}
