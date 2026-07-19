package api

import (
	"context"
	"fmt"
)

func validateEphemeralRef(r EphemeralMessageRef) error {
	if r.ChatID == nil {
		return fmt.Errorf("hermes: ephemeral chat_id is required")
	}
	if r.ReceiverUserID == 0 {
		return fmt.Errorf("hermes: ephemeral receiver_user_id is required")
	}
	if r.EphemeralMessageID == 0 {
		return fmt.Errorf("hermes: ephemeral_message_id is required")
	}
	return nil
}

type EditEphemeralMessageTextParams struct {
	ChatID             any                   `json:"chat_id"`
	ReceiverUserID     int64                 `json:"receiver_user_id"`
	EphemeralMessageID int                   `json:"ephemeral_message_id"`
	Text               string                `json:"text"`
	ParseMode          string                `json:"parse_mode,omitempty"`
	Entities           []MessageEntity       `json:"entities,omitempty"`
	LinkPreviewOptions *LinkPreviewOptions   `json:"link_preview_options,omitempty"`
	ReplyMarkup        *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (b *Client) EditEphemeralText(ctx context.Context, params EditEphemeralMessageTextParams) error {
	if err := validateEphemeralRef(EphemeralMessageRef{
		ChatID:             params.ChatID,
		ReceiverUserID:     params.ReceiverUserID,
		EphemeralMessageID: params.EphemeralMessageID,
	}); err != nil {
		return err
	}
	if params.Text == "" {
		return fmt.Errorf("hermes: ephemeral text is required")
	}
	return b.callTrue(ctx, "editEphemeralMessageText", params)
}

type InputMedia struct {
	Type       string          `json:"type"`
	Media      string          `json:"media"`
	Caption    string          `json:"caption,omitempty"`
	ParseMode  string          `json:"parse_mode,omitempty"`
	Entities   []MessageEntity `json:"caption_entities,omitempty"`
	HasSpoiler bool            `json:"has_spoiler,omitempty"`
}

type EditEphemeralMessageMediaParams struct {
	ChatID             any                   `json:"chat_id"`
	ReceiverUserID     int64                 `json:"receiver_user_id"`
	EphemeralMessageID int                   `json:"ephemeral_message_id"`
	Media              InputMedia            `json:"media"`
	ReplyMarkup        *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (b *Client) EditEphemeralMedia(ctx context.Context, params EditEphemeralMessageMediaParams) error {
	if err := validateEphemeralRef(EphemeralMessageRef{
		ChatID:             params.ChatID,
		ReceiverUserID:     params.ReceiverUserID,
		EphemeralMessageID: params.EphemeralMessageID,
	}); err != nil {
		return err
	}
	if params.Media.Type == "" || params.Media.Media == "" {
		return fmt.Errorf("hermes: ephemeral media type and media are required")
	}
	return b.callTrue(ctx, "editEphemeralMessageMedia", params)
}

type EditEphemeralMessageCaptionParams struct {
	ChatID             any                   `json:"chat_id"`
	ReceiverUserID     int64                 `json:"receiver_user_id"`
	EphemeralMessageID int                   `json:"ephemeral_message_id"`
	Caption            string                `json:"caption,omitempty"`
	ParseMode          string                `json:"parse_mode,omitempty"`
	CaptionEntities    []MessageEntity       `json:"caption_entities,omitempty"`
	ReplyMarkup        *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (b *Client) EditEphemeralCaption(ctx context.Context, params EditEphemeralMessageCaptionParams) error {
	if err := validateEphemeralRef(EphemeralMessageRef{
		ChatID:             params.ChatID,
		ReceiverUserID:     params.ReceiverUserID,
		EphemeralMessageID: params.EphemeralMessageID,
	}); err != nil {
		return err
	}
	return b.callTrue(ctx, "editEphemeralMessageCaption", params)
}

type EditEphemeralMessageReplyMarkupParams struct {
	ChatID             any                   `json:"chat_id"`
	ReceiverUserID     int64                 `json:"receiver_user_id"`
	EphemeralMessageID int                   `json:"ephemeral_message_id"`
	ReplyMarkup        *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (b *Client) EditEphemeralReplyMarkup(ctx context.Context, params EditEphemeralMessageReplyMarkupParams) error {
	if err := validateEphemeralRef(EphemeralMessageRef{
		ChatID:             params.ChatID,
		ReceiverUserID:     params.ReceiverUserID,
		EphemeralMessageID: params.EphemeralMessageID,
	}); err != nil {
		return err
	}
	return b.callTrue(ctx, "editEphemeralMessageReplyMarkup", params)
}

type DeleteEphemeralMessageParams struct {
	ChatID             any   `json:"chat_id"`
	ReceiverUserID     int64 `json:"receiver_user_id"`
	EphemeralMessageID int   `json:"ephemeral_message_id"`
}

func (b *Client) DeleteEphemeral(ctx context.Context, params DeleteEphemeralMessageParams) error {
	if err := validateEphemeralRef(EphemeralMessageRef{
		ChatID:             params.ChatID,
		ReceiverUserID:     params.ReceiverUserID,
		EphemeralMessageID: params.EphemeralMessageID,
	}); err != nil {
		return err
	}
	return b.callTrue(ctx, "deleteEphemeralMessage", params)
}

func (b *Client) callTrue(ctx context.Context, method string, params any) error {
	var ok bool
	if err := b.Call(ctx, method, params, &ok); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("hermes: %s returned false", method)
	}
	return nil
}
