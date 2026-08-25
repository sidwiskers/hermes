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
	Text               string                `json:"text,omitempty"`
	ParseMode          string                `json:"parse_mode,omitempty"`
	Entities           []MessageEntity       `json:"entities,omitempty"`
	RichMessage        *InputRichMessage     `json:"rich_message,omitempty"`
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
	if (params.Text == "") == (params.RichMessage == nil) {
		return fmt.Errorf("hermes: editEphemeralMessageText requires exactly one of text or rich_message")
	}
	if params.RichMessage != nil {
		if err := validateRichMessage(*params.RichMessage, false); err != nil {
			return err
		}
		if err := validateAttachmentUploads(*params.RichMessage, nil, "editEphemeralMessageText"); err != nil {
			return err
		}
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
	ChatID             any        `json:"chat_id"`
	ReceiverUserID     int64      `json:"receiver_user_id"`
	EphemeralMessageID int        `json:"ephemeral_message_id"`
	Media              InputMedia `json:"media"`
	// TypedMedia selects a discriminator-aware media variant. It is mutually
	// exclusive with the source-compatible Media field.
	TypedMedia  InputEditableMedia    `json:"-"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func editEphemeralMedia(params EditEphemeralMessageMediaParams) (InputEditableMedia, error) {
	legacyPresent := params.Media.Type != "" || params.Media.Media != ""
	if params.TypedMedia != nil {
		if legacyPresent {
			return nil, fmt.Errorf("hermes: editEphemeralMessageMedia media and typed_media are mutually exclusive")
		}
		return params.TypedMedia, nil
	}
	if !legacyPresent {
		return nil, fmt.Errorf("hermes: ephemeral media is required")
	}
	return params.Media, nil
}

func editEphemeralMediaParams(params EditEphemeralMessageMediaParams, media InputEditableMedia) any {
	return struct {
		ChatID             any                   `json:"chat_id"`
		ReceiverUserID     int64                 `json:"receiver_user_id"`
		EphemeralMessageID int                   `json:"ephemeral_message_id"`
		Media              InputEditableMedia    `json:"media"`
		ReplyMarkup        *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	}{
		ChatID:             params.ChatID,
		ReceiverUserID:     params.ReceiverUserID,
		EphemeralMessageID: params.EphemeralMessageID,
		Media:              media,
		ReplyMarkup:        params.ReplyMarkup,
	}
}

func (b *Client) EditEphemeralMedia(ctx context.Context, params EditEphemeralMessageMediaParams) error {
	if err := validateEphemeralRef(EphemeralMessageRef{
		ChatID:             params.ChatID,
		ReceiverUserID:     params.ReceiverUserID,
		EphemeralMessageID: params.EphemeralMessageID,
	}); err != nil {
		return err
	}
	media, err := editEphemeralMedia(params)
	if err != nil {
		return err
	}
	if media.editableMediaSource() == "" {
		return fmt.Errorf("hermes: ephemeral media is required")
	}
	if generic, ok := media.(InputMedia); ok && generic.Type == "" {
		return fmt.Errorf("hermes: generic ephemeral media type is required")
	}
	if livePhoto, ok := media.(InputMediaLivePhoto); ok && livePhoto.Photo == "" {
		return fmt.Errorf("hermes: ephemeral live photo requires photo")
	}
	if err := validateAttachmentUploads(media, nil, "editEphemeralMessageMedia"); err != nil {
		return err
	}
	return b.callTrue(ctx, "editEphemeralMessageMedia", editEphemeralMediaParams(params, media))
}

// EditEphemeralMediaUpload streams every attach:// reference in Media.
func (b *Client) EditEphemeralMediaUpload(ctx context.Context, params EditEphemeralMessageMediaParams, uploads ...Upload) error {
	if len(uploads) == 0 {
		return b.EditEphemeralMedia(ctx, params)
	}
	if err := validateEphemeralRef(EphemeralMessageRef{
		ChatID:             params.ChatID,
		ReceiverUserID:     params.ReceiverUserID,
		EphemeralMessageID: params.EphemeralMessageID,
	}); err != nil {
		return err
	}
	media, err := editEphemeralMedia(params)
	if err != nil {
		return err
	}
	if media.editableMediaSource() == "" {
		return fmt.Errorf("hermes: ephemeral media is required")
	}
	if err := validateAttachmentUploads(media, uploads, "editEphemeralMessageMedia"); err != nil {
		return err
	}
	fields, err := newFormFields(params.ChatID)
	if err != nil {
		return err
	}
	fields.Int64("receiver_user_id", params.ReceiverUserID)
	fields.Int("ephemeral_message_id", params.EphemeralMessageID)
	if err = fields.JSON("media", media); err != nil {
		return err
	}
	if params.ReplyMarkup != nil {
		if err = fields.JSON("reply_markup", params.ReplyMarkup); err != nil {
			return err
		}
	}
	var ok bool
	if err = b.CallMultipart(ctx, "editEphemeralMessageMedia", fields, uploads, &ok); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("hermes: editEphemeralMessageMedia returned false")
	}
	return nil
}

type EditEphemeralMessageCaptionParams struct {
	ChatID                any                   `json:"chat_id"`
	ReceiverUserID        int64                 `json:"receiver_user_id"`
	EphemeralMessageID    int                   `json:"ephemeral_message_id"`
	Caption               string                `json:"caption,omitempty"`
	ParseMode             string                `json:"parse_mode,omitempty"`
	CaptionEntities       []MessageEntity       `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia bool                  `json:"show_caption_above_media,omitempty"`
	ReplyMarkup           *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
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
