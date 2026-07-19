package api

import (
	"context"
	"fmt"
	"strings"
)

type SendGameParams struct {
	BusinessConnectionID string                `json:"business_connection_id,omitempty"`
	ChatID               any                   `json:"chat_id"`
	MessageThreadID      int                   `json:"message_thread_id,omitempty"`
	GameShortName        string                `json:"game_short_name"`
	DisableNotification  bool                  `json:"disable_notification,omitempty"`
	ProtectContent       bool                  `json:"protect_content,omitempty"`
	AllowPaidBroadcast   bool                  `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID      string                `json:"message_effect_id,omitempty"`
	ReplyParameters      *ReplyParameters      `json:"reply_parameters,omitempty"`
	ReplyMarkup          *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (client *Client) SendGame(ctx context.Context, params SendGameParams) (*Message, error) {
	if err := validateChatID(params.ChatID, "sendGame"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(params.GameShortName) == "" {
		return nil, fmt.Errorf("hermes: sendGame game_short_name is required")
	}
	return callMessage(ctx, client, "sendGame", params)
}

type SetGameScoreParams struct {
	UserID             int64  `json:"user_id"`
	Score              int    `json:"score"`
	Force              bool   `json:"force,omitempty"`
	DisableEditMessage bool   `json:"disable_edit_message,omitempty"`
	ChatID             int64  `json:"chat_id,omitempty"`
	MessageID          int    `json:"message_id,omitempty"`
	InlineMessageID    string `json:"inline_message_id,omitempty"`
}

func (client *Client) SetGameScore(ctx context.Context, params SetGameScoreParams) (*Message, error) {
	if params.UserID == 0 || params.Score < 0 {
		return nil, fmt.Errorf("hermes: setGameScore requires user_id and a non-negative score")
	}
	if err := validateGameMessageTarget(params.ChatID, params.MessageID, params.InlineMessageID, "setGameScore"); err != nil {
		return nil, err
	}
	return callMessageOrBool(ctx, client, "setGameScore", params)
}

type GetGameHighScoresParams struct {
	UserID          int64  `json:"user_id"`
	ChatID          int64  `json:"chat_id,omitempty"`
	MessageID       int    `json:"message_id,omitempty"`
	InlineMessageID string `json:"inline_message_id,omitempty"`
}

func (client *Client) GetGameHighScores(ctx context.Context, params GetGameHighScoresParams) ([]GameHighScore, error) {
	if params.UserID == 0 {
		return nil, fmt.Errorf("hermes: getGameHighScores user_id is required")
	}
	if err := validateGameMessageTarget(params.ChatID, params.MessageID, params.InlineMessageID, "getGameHighScores"); err != nil {
		return nil, err
	}
	return Call[[]GameHighScore](ctx, client, "getGameHighScores", params)
}

func validateGameMessageTarget(chatID int64, messageID int, inlineMessageID, method string) error {
	if strings.TrimSpace(inlineMessageID) != "" {
		if chatID != 0 || messageID != 0 {
			return fmt.Errorf("hermes: %s inline_message_id cannot be combined with chat_id or message_id", method)
		}
		return nil
	}
	if chatID == 0 || messageID == 0 {
		return fmt.Errorf("hermes: %s chat_id and message_id are required", method)
	}
	return nil
}
