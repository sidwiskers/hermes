package api

import (
	"context"
	"fmt"
	"strings"
)

func validateInlineQueryResult(result InlineQueryResult) error {
	if isNilUnion(result) || strings.TrimSpace(result.InlineQueryResultType()) == "" || strings.TrimSpace(result.InlineQueryResultID()) == "" {
		return fmt.Errorf("hermes: inline query result type and id are required")
	}
	if len(result.InlineQueryResultID()) > 64 {
		return fmt.Errorf("hermes: inline query result id must not exceed 64 bytes")
	}
	return nil
}

type AnswerInlineQueryParams struct {
	InlineQueryID string                    `json:"inline_query_id"`
	Results       []InlineQueryResult       `json:"results"`
	CacheTime     int                       `json:"cache_time,omitempty"`
	IsPersonal    bool                      `json:"is_personal,omitempty"`
	NextOffset    string                    `json:"next_offset,omitempty"`
	Button        *InlineQueryResultsButton `json:"button,omitempty"`
}

func (client *Client) AnswerInlineQuery(ctx context.Context, params AnswerInlineQueryParams) error {
	if strings.TrimSpace(params.InlineQueryID) == "" {
		return fmt.Errorf("hermes: answerInlineQuery inline_query_id is required")
	}
	if len(params.Results) > 50 {
		return fmt.Errorf("hermes: answerInlineQuery accepts at most 50 results")
	}
	if len(params.NextOffset) > 64 {
		return fmt.Errorf("hermes: answerInlineQuery next_offset must not exceed 64 bytes")
	}
	for _, result := range params.Results {
		if err := validateInlineQueryResult(result); err != nil {
			return err
		}
	}
	if params.Button != nil {
		hasWebApp := params.Button.WebApp != nil
		hasStart := params.Button.StartParameter != ""
		if strings.TrimSpace(params.Button.Text) == "" || hasWebApp == hasStart {
			return fmt.Errorf("hermes: inline query button requires text and exactly one action")
		}
	}
	return client.callTrue(ctx, "answerInlineQuery", params)
}

type AnswerWebAppQueryParams struct {
	WebAppQueryID string            `json:"web_app_query_id"`
	Result        InlineQueryResult `json:"result"`
}

func (client *Client) AnswerWebAppQuery(ctx context.Context, params AnswerWebAppQueryParams) (SentWebAppMessage, error) {
	if strings.TrimSpace(params.WebAppQueryID) == "" {
		return SentWebAppMessage{}, fmt.Errorf("hermes: answerWebAppQuery web_app_query_id is required")
	}
	if err := validateInlineQueryResult(params.Result); err != nil {
		return SentWebAppMessage{}, err
	}
	return Call[SentWebAppMessage](ctx, client, "answerWebAppQuery", params)
}

type SavePreparedInlineMessageParams struct {
	UserID            int64             `json:"user_id"`
	Result            InlineQueryResult `json:"result"`
	AllowUserChats    bool              `json:"allow_user_chats,omitempty"`
	AllowBotChats     bool              `json:"allow_bot_chats,omitempty"`
	AllowGroupChats   bool              `json:"allow_group_chats,omitempty"`
	AllowChannelChats bool              `json:"allow_channel_chats,omitempty"`
}

func (client *Client) SavePreparedInlineMessage(ctx context.Context, params SavePreparedInlineMessageParams) (PreparedInlineMessage, error) {
	if params.UserID == 0 {
		return PreparedInlineMessage{}, fmt.Errorf("hermes: savePreparedInlineMessage user_id is required")
	}
	if err := validateInlineQueryResult(params.Result); err != nil {
		return PreparedInlineMessage{}, err
	}
	return Call[PreparedInlineMessage](ctx, client, "savePreparedInlineMessage", params)
}

type SavePreparedKeyboardButtonParams struct {
	UserID int64          `json:"user_id"`
	Button KeyboardButton `json:"button"`
}

func (client *Client) SavePreparedKeyboardButton(ctx context.Context, params SavePreparedKeyboardButtonParams) (PreparedKeyboardButton, error) {
	if params.UserID == 0 {
		return PreparedKeyboardButton{}, fmt.Errorf("hermes: savePreparedKeyboardButton user_id is required")
	}
	actions := 0
	if params.Button.RequestUsers != nil {
		actions++
	}
	if params.Button.RequestChat != nil {
		actions++
	}
	if params.Button.RequestManagedBot != nil {
		actions++
	}
	if strings.TrimSpace(params.Button.Text) == "" || actions != 1 {
		return PreparedKeyboardButton{}, fmt.Errorf("hermes: prepared keyboard button requires text and exactly one supported request action")
	}
	return Call[PreparedKeyboardButton](ctx, client, "savePreparedKeyboardButton", params)
}

type AnswerGuestQueryParams struct {
	GuestQueryID string            `json:"guest_query_id"`
	Result       InlineQueryResult `json:"result"`
}

func (client *Client) AnswerGuestQuery(ctx context.Context, params AnswerGuestQueryParams) (SentGuestMessage, error) {
	if strings.TrimSpace(params.GuestQueryID) == "" {
		return SentGuestMessage{}, fmt.Errorf("hermes: answerGuestQuery guest_query_id is required")
	}
	if err := validateInlineQueryResult(params.Result); err != nil {
		return SentGuestMessage{}, err
	}
	return Call[SentGuestMessage](ctx, client, "answerGuestQuery", params)
}
