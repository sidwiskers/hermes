package api

import (
	"context"
	"io"
)

type SendVideoParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	Video                   string                   `json:"video"`
	Duration                int                      `json:"duration,omitempty"`
	Width                   int                      `json:"width,omitempty"`
	Height                  int                      `json:"height,omitempty"`
	Thumbnail               string                   `json:"thumbnail,omitempty"`
	Cover                   string                   `json:"cover,omitempty"`
	StartTimestamp          int                      `json:"start_timestamp,omitempty"`
	Caption                 string                   `json:"caption,omitempty"`
	ParseMode               string                   `json:"parse_mode,omitempty"`
	CaptionEntities         []MessageEntity          `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia   bool                     `json:"show_caption_above_media,omitempty"`
	HasSpoiler              bool                     `json:"has_spoiler,omitempty"`
	SupportsStreaming       bool                     `json:"supports_streaming,omitempty"`
	DisableNotification     bool                     `json:"disable_notification,omitempty"`
	ProtectContent          bool                     `json:"protect_content,omitempty"`
	AllowPaidBroadcast      bool                     `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID         string                   `json:"message_effect_id,omitempty"`
	SuggestedPostParameters *SuggestedPostParameters `json:"suggested_post_parameters,omitempty"`
	ReplyParameters         *ReplyParameters         `json:"reply_parameters,omitempty"`
	ReplyMarkup             ReplyMarkup              `json:"reply_markup,omitempty"`
	ReceiverUserID          int64                    `json:"receiver_user_id,omitempty"`
	CallbackQueryID         string                   `json:"callback_query_id,omitempty"`
}

func (p SendVideoParams) base() SendBaseParams {
	return SendBaseParams{
		BusinessConnectionID:    p.BusinessConnectionID,
		ChatID:                  p.ChatID,
		MessageThreadID:         p.MessageThreadID,
		DirectMessagesTopicID:   p.DirectMessagesTopicID,
		DisableNotification:     p.DisableNotification,
		ProtectContent:          p.ProtectContent,
		AllowPaidBroadcast:      p.AllowPaidBroadcast,
		MessageEffectID:         p.MessageEffectID,
		SuggestedPostParameters: p.SuggestedPostParameters,
		ReplyParameters:         p.ReplyParameters,
		ReplyMarkup:             p.ReplyMarkup,
		ReceiverUserID:          p.ReceiverUserID,
		CallbackQueryID:         p.CallbackQueryID,
	}
}
func (p SendVideoParams) caption() CaptionParams {
	return CaptionParams{
		Caption:               p.Caption,
		ParseMode:             p.ParseMode,
		CaptionEntities:       p.CaptionEntities,
		ShowCaptionAboveMedia: p.ShowCaptionAboveMedia,
	}
}
func (b *Client) SendVideo(ctx context.Context, p SendVideoParams) (*Message, error) {
	return b.sendMediaJSON(ctx, "sendVideo", p.ChatID, p.Video, p)
}
func (b *Client) SendVideoUpload(ctx context.Context, p SendVideoParams, name string, reader io.Reader) (*Message, error) {
	fields, err := mediaFields(p.base(), p.caption())
	if err != nil {
		return nil, err
	}
	fields.Int("duration", p.Duration)
	fields.Int("width", p.Width)
	fields.Int("height", p.Height)
	fields.String("thumbnail", p.Thumbnail)
	fields.String("cover", p.Cover)
	fields.Int("start_timestamp", p.StartTimestamp)
	fields.Bool("has_spoiler", p.HasSpoiler)
	fields.Bool("supports_streaming", p.SupportsStreaming)
	return b.sendUpload(ctx, "sendVideo", "video", fields, name, reader)
}

type SendVideoNoteParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	VideoNote               string                   `json:"video_note"`
	Duration                int                      `json:"duration,omitempty"`
	Length                  int                      `json:"length,omitempty"`
	Thumbnail               string                   `json:"thumbnail,omitempty"`
	DisableNotification     bool                     `json:"disable_notification,omitempty"`
	ProtectContent          bool                     `json:"protect_content,omitempty"`
	AllowPaidBroadcast      bool                     `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID         string                   `json:"message_effect_id,omitempty"`
	SuggestedPostParameters *SuggestedPostParameters `json:"suggested_post_parameters,omitempty"`
	ReplyParameters         *ReplyParameters         `json:"reply_parameters,omitempty"`
	ReplyMarkup             ReplyMarkup              `json:"reply_markup,omitempty"`
	ReceiverUserID          int64                    `json:"receiver_user_id,omitempty"`
	CallbackQueryID         string                   `json:"callback_query_id,omitempty"`
}

func (p SendVideoNoteParams) base() SendBaseParams {
	return SendBaseParams{
		BusinessConnectionID:    p.BusinessConnectionID,
		ChatID:                  p.ChatID,
		MessageThreadID:         p.MessageThreadID,
		DirectMessagesTopicID:   p.DirectMessagesTopicID,
		DisableNotification:     p.DisableNotification,
		ProtectContent:          p.ProtectContent,
		AllowPaidBroadcast:      p.AllowPaidBroadcast,
		MessageEffectID:         p.MessageEffectID,
		SuggestedPostParameters: p.SuggestedPostParameters,
		ReplyParameters:         p.ReplyParameters,
		ReplyMarkup:             p.ReplyMarkup,
		ReceiverUserID:          p.ReceiverUserID,
		CallbackQueryID:         p.CallbackQueryID,
	}
}
func (b *Client) SendVideoNote(ctx context.Context, p SendVideoNoteParams) (*Message, error) {
	return b.sendMediaJSON(ctx, "sendVideoNote", p.ChatID, p.VideoNote, p)
}
func (b *Client) SendVideoNoteUpload(ctx context.Context, p SendVideoNoteParams, name string, reader io.Reader) (*Message, error) {
	fields, err := newFormFields(p.ChatID)
	if err != nil {
		return nil, err
	}
	if err = addSendBaseFields(fields, p.base()); err != nil {
		return nil, err
	}
	fields.Int("duration", p.Duration)
	fields.Int("length", p.Length)
	fields.String("thumbnail", p.Thumbnail)
	return b.sendUpload(ctx, "sendVideoNote", "video_note", fields, name, reader)
}
