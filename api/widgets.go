package api

import (
	"bytes"
	"compress/zlib"
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	maxWidgetHTMLBytes = 32 * 1024
	maxWidgetHTMLZlibBytes = 4 * 1024
	maxWidgetStateBytes = 1024
)

func validateWidgetState(value, field string) error {
	if len(value) > maxWidgetStateBytes {
		return fmt.Errorf("hermes: %s must not exceed %d bytes", field, maxWidgetStateBytes)
	}
	if !utf8.ValidString(value) || strings.IndexByte(value, 0) >= 0 {
		return fmt.Errorf("hermes: %s must be valid UTF-8 without NUL bytes", field)
	}
	return nil
}

func validateWidgetHTML(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("hermes: sendWidget html is required")
	}
	if len(value) > maxWidgetHTMLBytes {
		return fmt.Errorf("hermes: sendWidget html must not exceed %d bytes", maxWidgetHTMLBytes)
	}
	if !utf8.ValidString(value) || strings.IndexByte(value, 0) >= 0 {
		return fmt.Errorf("hermes: sendWidget html must be valid UTF-8 without NUL bytes")
	}

	var compressed bytes.Buffer
	writer, err := zlib.NewWriterLevel(&compressed, zlib.BestCompression)
	if err != nil {
		return fmt.Errorf("hermes: initialize widget HTML compression: %w", err)
	}
	if _, err := writer.Write([]byte(value)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("hermes: compress widget HTML: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("hermes: finish widget HTML compression: %w", err)
	}
	if compressed.Len() > maxWidgetHTMLZlibBytes {
		return fmt.Errorf("hermes: sendWidget compressed html must not exceed %d bytes", maxWidgetHTMLZlibBytes)
	}
	return nil
}

// SendWidgetParams sends a standalone HTML widget message.
type SendWidgetParams struct {
	ChatID                any              `json:"chat_id"`
	MessageThreadID       int              `json:"message_thread_id,omitempty"`
	HTML                  string           `json:"html"`
	Width                 int              `json:"width,omitempty"`
	Height                int              `json:"height,omitempty"`
	PublicData            string           `json:"public_data,omitempty"`
	ChatData              string           `json:"chat_data,omitempty"`
	Caption               string           `json:"caption,omitempty"`
	ParseMode             string           `json:"parse_mode,omitempty"`
	CaptionEntities       []MessageEntity  `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia bool             `json:"show_caption_above_media,omitempty"`
	AllowQueries          bool             `json:"allow_queries,omitempty"`
	IsShared              bool             `json:"is_shared,omitempty"`
	DisableNotification   bool             `json:"disable_notification,omitempty"`
	ProtectContent        bool             `json:"protect_content,omitempty"`
	AllowPaidBroadcast    bool             `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID       string           `json:"message_effect_id,omitempty"`
	ReplyParameters       *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup           ReplyMarkup      `json:"reply_markup,omitempty"`
}

// SendWidget sends an HTML widget using Telegram's native widget renderer.
func (client *Client) SendWidget(ctx context.Context, params SendWidgetParams) (*Message, error) {
	if err := validateChatID(params.ChatID, "sendWidget"); err != nil {
		return nil, err
	}
	if err := validateWidgetHTML(params.HTML); err != nil {
		return nil, err
	}
	if err := validateWidgetState(params.PublicData, "sendWidget public_data"); err != nil {
		return nil, err
	}
	if err := validateWidgetState(params.ChatData, "sendWidget chat_data"); err != nil {
		return nil, err
	}
	return callMessage(ctx, client, "sendWidget", params)
}

// SendSharedWidgetParams sends a previously registered shared widget.
type SendSharedWidgetParams struct {
	ChatID                any              `json:"chat_id"`
	MessageThreadID       int              `json:"message_thread_id,omitempty"`
	WidgetID              string           `json:"widget_id"`
	ChatData              string           `json:"chat_data,omitempty"`
	Caption               string           `json:"caption,omitempty"`
	ParseMode             string           `json:"parse_mode,omitempty"`
	CaptionEntities       []MessageEntity  `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia bool             `json:"show_caption_above_media,omitempty"`
	DisableNotification   bool             `json:"disable_notification,omitempty"`
	ProtectContent        bool             `json:"protect_content,omitempty"`
	AllowPaidBroadcast    bool             `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID       string           `json:"message_effect_id,omitempty"`
	ReplyParameters       *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup           ReplyMarkup      `json:"reply_markup,omitempty"`
}

// SendSharedWidget sends a reusable widget registered with Telegram.
func (client *Client) SendSharedWidget(ctx context.Context, params SendSharedWidgetParams) (*Message, error) {
	if err := validateChatID(params.ChatID, "sendSharedWidget"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(params.WidgetID) == "" {
		return nil, fmt.Errorf("hermes: sendSharedWidget widget_id is required")
	}
	if err := validateWidgetState(params.ChatData, "sendSharedWidget chat_data"); err != nil {
		return nil, err
	}
	return callMessage(ctx, client, "sendSharedWidget", params)
}

type AnswerWidgetQueryParams struct {
	WidgetQueryID string `json:"widget_query_id"`
	Data          string `json:"data,omitempty"`
}

func (client *Client) AnswerWidgetQuery(ctx context.Context, params AnswerWidgetQueryParams) error {
	if strings.TrimSpace(params.WidgetQueryID) == "" {
		return fmt.Errorf("hermes: answerWidgetQuery widget_query_id is required")
	}
	if err := validateWidgetState(params.Data, "answerWidgetQuery data"); err != nil {
		return err
	}
	return client.callTrue(ctx, "answerWidgetQuery", params)
}

type EditMessageWidgetDataParams struct {
	ChatID          any     `json:"chat_id,omitempty"`
	MessageID       int     `json:"message_id,omitempty"`
	InlineMessageID string  `json:"inline_message_id,omitempty"`
	PublicData      *string `json:"public_data,omitempty"`
	ChatData        *string `json:"chat_data,omitempty"`
}

func (client *Client) EditMessageWidgetData(ctx context.Context, params EditMessageWidgetDataParams) error {
	if err := validateMessageEditTarget(params.ChatID, params.MessageID, params.InlineMessageID, "editMessageWidgetData"); err != nil {
		return err
	}
	if params.PublicData != nil {
		if err := validateWidgetState(*params.PublicData, "editMessageWidgetData public_data"); err != nil {
			return err
		}
	}
	if params.ChatData != nil {
		if err := validateWidgetState(*params.ChatData, "editMessageWidgetData chat_data"); err != nil {
			return err
		}
	}
	return client.callTrue(ctx, "editMessageWidgetData", params)
}

type EditSharedWidgetDataParams struct {
	WidgetID   string `json:"widget_id"`
	PublicData string `json:"public_data"`
	CacheTime  int    `json:"cache_time,omitempty"`
}

func (client *Client) EditSharedWidgetData(ctx context.Context, params EditSharedWidgetDataParams) error {
	if strings.TrimSpace(params.WidgetID) == "" {
		return fmt.Errorf("hermes: editSharedWidgetData widget_id is required")
	}
	if err := validateWidgetState(params.PublicData, "editSharedWidgetData public_data"); err != nil {
		return err
	}
	if params.CacheTime != 0 && (params.CacheTime < 5 || params.CacheTime > 3600) {
		return fmt.Errorf("hermes: editSharedWidgetData cache_time must be between 5 and 3600 seconds")
	}
	return client.callTrue(ctx, "editSharedWidgetData", params)
}

type GetUniqueGiftParams struct {
	Name string `json:"name"`
}

func (client *Client) GetUniqueGift(ctx context.Context, params GetUniqueGiftParams) (UniqueGift, error) {
	if strings.TrimSpace(params.Name) == "" {
		return UniqueGift{}, fmt.Errorf("hermes: getUniqueGift name is required")
	}
	return Call[UniqueGift](ctx, client, "getUniqueGift", params)
}

type DeleteBusinessAccountGiftParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
	OwnedGiftID          string `json:"owned_gift_id"`
}

func (client *Client) DeleteBusinessAccountGift(ctx context.Context, params DeleteBusinessAccountGiftParams) error {
	if err := validateOwnedGift(OwnedGiftParams{
		BusinessConnectionID: params.BusinessConnectionID,
		OwnedGiftID:          params.OwnedGiftID,
	}, "deleteBusinessAccountGift"); err != nil {
		return err
	}
	return client.callTrue(ctx, "deleteBusinessAccountGift", params)
}

func (client *Client) GetChatShareThemeBackdrops(ctx context.Context) (ChatShareThemeBackdrops, error) {
	return Call[ChatShareThemeBackdrops](ctx, client, "getChatShareThemeBackdrops", nil)
}
