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
