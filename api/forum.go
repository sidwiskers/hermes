package api

import (
	"context"
	"fmt"
	"unicode/utf8"
)

const (
	ForumIconBlue   = 0x6FB9F0
	ForumIconYellow = 0xFFD67E
	ForumIconPurple = 0xCB86DB
	ForumIconGreen  = 0x8EEE98
	ForumIconPink   = 0xFF93B2
	ForumIconRed    = 0xFB6F5F
)

func validForumIconColor(color int) bool {
	switch color {
	case 0, ForumIconBlue, ForumIconYellow, ForumIconPurple, ForumIconGreen, ForumIconPink, ForumIconRed:
		return true
	default:
		return false
	}
}

func validateTopicName(name string, required bool) error {
	length := utf8.RuneCountInString(name)
	if required && length == 0 {
		return fmt.Errorf("hermes: forum topic name is required")
	}
	if length > 128 {
		return fmt.Errorf("hermes: forum topic name must not exceed 128 characters")
	}
	return nil
}

func (client *Client) GetForumTopicIconStickers(ctx context.Context) ([]Sticker, error) {
	return Call[[]Sticker](ctx, client, "getForumTopicIconStickers", nil)
}

type CreateForumTopicParams struct {
	ChatID            any    `json:"chat_id"`
	Name              string `json:"name"`
	IconColor         int    `json:"icon_color,omitempty"`
	IconCustomEmojiID string `json:"icon_custom_emoji_id,omitempty"`
}

func (client *Client) CreateForumTopic(ctx context.Context, params CreateForumTopicParams) (ForumTopic, error) {
	if err := validateChatID(params.ChatID, "createForumTopic"); err != nil {
		return ForumTopic{}, err
	}
	if err := validateTopicName(params.Name, true); err != nil {
		return ForumTopic{}, err
	}
	if !validForumIconColor(params.IconColor) {
		return ForumTopic{}, fmt.Errorf("hermes: createForumTopic icon_color is unsupported")
	}
	return Call[ForumTopic](ctx, client, "createForumTopic", params)
}

type EditForumTopicParams struct {
	ChatID            any     `json:"chat_id"`
	MessageThreadID   int     `json:"message_thread_id"`
	Name              string  `json:"name,omitempty"`
	IconCustomEmojiID *string `json:"icon_custom_emoji_id,omitempty"`
}

func (client *Client) EditForumTopic(ctx context.Context, params EditForumTopicParams) error {
	if err := validateForumTopicTarget(params.ChatID, params.MessageThreadID, "editForumTopic"); err != nil {
		return err
	}
	if err := validateTopicName(params.Name, false); err != nil {
		return err
	}
	return client.callTrue(ctx, "editForumTopic", params)
}

type ForumTopicTargetParams struct {
	ChatID          any `json:"chat_id"`
	MessageThreadID int `json:"message_thread_id"`
}

func validateForumTopicTarget(chatID any, threadID int, method string) error {
	if err := validateChatID(chatID, method); err != nil {
		return err
	}
	if threadID == 0 {
		return fmt.Errorf("hermes: %s message_thread_id is required", method)
	}
	return nil
}

func (client *Client) forumTopicAction(ctx context.Context, method string, params ForumTopicTargetParams) error {
	if err := validateForumTopicTarget(params.ChatID, params.MessageThreadID, method); err != nil {
		return err
	}
	return client.callTrue(ctx, method, params)
}

func (client *Client) CloseForumTopic(ctx context.Context, params ForumTopicTargetParams) error {
	return client.forumTopicAction(ctx, "closeForumTopic", params)
}

func (client *Client) ReopenForumTopic(ctx context.Context, params ForumTopicTargetParams) error {
	return client.forumTopicAction(ctx, "reopenForumTopic", params)
}

func (client *Client) DeleteForumTopic(ctx context.Context, params ForumTopicTargetParams) error {
	return client.forumTopicAction(ctx, "deleteForumTopic", params)
}

func (client *Client) UnpinAllForumTopicMessages(ctx context.Context, params ForumTopicTargetParams) error {
	return client.forumTopicAction(ctx, "unpinAllForumTopicMessages", params)
}

type EditGeneralForumTopicParams struct {
	ChatID any    `json:"chat_id"`
	Name   string `json:"name"`
}

func (client *Client) EditGeneralForumTopic(ctx context.Context, params EditGeneralForumTopicParams) error {
	if err := validateChatID(params.ChatID, "editGeneralForumTopic"); err != nil {
		return err
	}
	if err := validateTopicName(params.Name, true); err != nil {
		return err
	}
	return client.callTrue(ctx, "editGeneralForumTopic", params)
}

type GeneralForumTopicParams struct {
	ChatID any `json:"chat_id"`
}

func (client *Client) generalForumTopicAction(ctx context.Context, method string, params GeneralForumTopicParams) error {
	if err := validateChatID(params.ChatID, method); err != nil {
		return err
	}
	return client.callTrue(ctx, method, params)
}

func (client *Client) CloseGeneralForumTopic(ctx context.Context, params GeneralForumTopicParams) error {
	return client.generalForumTopicAction(ctx, "closeGeneralForumTopic", params)
}

func (client *Client) ReopenGeneralForumTopic(ctx context.Context, params GeneralForumTopicParams) error {
	return client.generalForumTopicAction(ctx, "reopenGeneralForumTopic", params)
}

func (client *Client) HideGeneralForumTopic(ctx context.Context, params GeneralForumTopicParams) error {
	return client.generalForumTopicAction(ctx, "hideGeneralForumTopic", params)
}

func (client *Client) UnhideGeneralForumTopic(ctx context.Context, params GeneralForumTopicParams) error {
	return client.generalForumTopicAction(ctx, "unhideGeneralForumTopic", params)
}

func (client *Client) UnpinAllGeneralForumTopicMessages(ctx context.Context, params GeneralForumTopicParams) error {
	return client.generalForumTopicAction(ctx, "unpinAllGeneralForumTopicMessages", params)
}
