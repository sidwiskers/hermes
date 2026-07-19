package api

import (
	"context"
	"fmt"
	"strings"
)

type SetMessageReactionParams struct {
	ChatID    any            `json:"chat_id"`
	MessageID int            `json:"message_id"`
	Reaction  []ReactionType `json:"reaction,omitempty"`
	IsBig     bool           `json:"is_big,omitempty"`
}

func (b *Client) SetMessageReaction(ctx context.Context, params SetMessageReactionParams) error {
	if err := validateMessageTarget(params.ChatID, params.MessageID, "setMessageReaction"); err != nil {
		return err
	}
	if len(params.Reaction) > 1 {
		return fmt.Errorf("hermes: bots can set at most one reaction per message")
	}
	for _, reaction := range params.Reaction {
		if err := validateReaction(reaction); err != nil {
			return err
		}
	}
	return b.callTrue(ctx, "setMessageReaction", params)
}

type DeleteMessageReactionParams struct {
	ChatID      any   `json:"chat_id"`
	MessageID   int   `json:"message_id"`
	UserID      int64 `json:"user_id,omitempty"`
	ActorChatID int64 `json:"actor_chat_id,omitempty"`
}

func (b *Client) DeleteMessageReaction(ctx context.Context, params DeleteMessageReactionParams) error {
	if err := validateMessageTarget(params.ChatID, params.MessageID, "deleteMessageReaction"); err != nil {
		return err
	}
	if err := validateReactionActor(params.UserID, params.ActorChatID, "deleteMessageReaction"); err != nil {
		return err
	}
	return b.callTrue(ctx, "deleteMessageReaction", params)
}

type DeleteAllMessageReactionsParams struct {
	ChatID      any   `json:"chat_id"`
	UserID      int64 `json:"user_id,omitempty"`
	ActorChatID int64 `json:"actor_chat_id,omitempty"`
}

func (b *Client) DeleteAllMessageReactions(ctx context.Context, params DeleteAllMessageReactionsParams) error {
	if err := validateChatID(params.ChatID, "deleteAllMessageReactions"); err != nil {
		return err
	}
	if err := validateReactionActor(params.UserID, params.ActorChatID, "deleteAllMessageReactions"); err != nil {
		return err
	}
	return b.callTrue(ctx, "deleteAllMessageReactions", params)
}

func validateReaction(reaction ReactionType) error {
	switch reaction.Type {
	case ReactionEmoji:
		if strings.TrimSpace(reaction.Emoji) == "" {
			return fmt.Errorf("hermes: emoji reaction requires emoji")
		}
	case ReactionCustomEmoji:
		if strings.TrimSpace(reaction.CustomEmojiID) == "" {
			return fmt.Errorf("hermes: custom emoji reaction requires custom_emoji_id")
		}
	case ReactionPaid:
		return fmt.Errorf("hermes: bots can't set paid reactions")
	default:
		return fmt.Errorf("hermes: unsupported reaction type %q", reaction.Type)
	}
	return nil
}

func validateReactionActor(userID, actorChatID int64, method string) error {
	if userID != 0 && actorChatID != 0 {
		return fmt.Errorf("hermes: %s user_id and actor_chat_id are mutually exclusive", method)
	}
	return nil
}

func validateMessageTarget(chatID any, messageID int, method string) error {
	if err := validateChatID(chatID, method); err != nil {
		return err
	}
	if messageID == 0 {
		return fmt.Errorf("hermes: %s message_id is required", method)
	}
	return nil
}
