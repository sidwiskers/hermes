package api

import (
	"context"
	"io"
)

type SendPhotoParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	Photo                   string                   `json:"photo"`
	Caption                 string                   `json:"caption,omitempty"`
	ParseMode               string                   `json:"parse_mode,omitempty"`
	CaptionEntities         []MessageEntity          `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia   bool                     `json:"show_caption_above_media,omitempty"`
	HasSpoiler              bool                     `json:"has_spoiler,omitempty"`
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

func (p SendPhotoParams) base() SendBaseParams {
	return SendBaseParams{
		BusinessConnectionID: p.BusinessConnectionID, ChatID: p.ChatID, MessageThreadID: p.MessageThreadID,
		DirectMessagesTopicID: p.DirectMessagesTopicID,
		DisableNotification:   p.DisableNotification, ProtectContent: p.ProtectContent,
		AllowPaidBroadcast: p.AllowPaidBroadcast, MessageEffectID: p.MessageEffectID,
		SuggestedPostParameters: p.SuggestedPostParameters,
		ReplyParameters:         p.ReplyParameters, ReplyMarkup: p.ReplyMarkup,
		ReceiverUserID: p.ReceiverUserID, CallbackQueryID: p.CallbackQueryID,
	}
}
func (p SendPhotoParams) caption() CaptionParams {
	return CaptionParams{
		Caption:               p.Caption,
		ParseMode:             p.ParseMode,
		CaptionEntities:       p.CaptionEntities,
		ShowCaptionAboveMedia: p.ShowCaptionAboveMedia,
	}
}

func (b *Client) SendPhoto(ctx context.Context, params SendPhotoParams) (*Message, error) {
	return b.sendMediaJSON(ctx, "sendPhoto", params.ChatID, params.Photo, params)
}
func (b *Client) SendPhotoUpload(ctx context.Context, params SendPhotoParams, name string, reader io.Reader) (*Message, error) {
	fields, err := mediaFields(params.base(), params.caption())
	if err != nil {
		return nil, err
	}
	fields.Bool("has_spoiler", params.HasSpoiler)
	return b.sendUpload(ctx, "sendPhoto", "photo", fields, name, reader)
}

type SendAnimationParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	Animation               string                   `json:"animation"`
	Duration                int                      `json:"duration,omitempty"`
	Width                   int                      `json:"width,omitempty"`
	Height                  int                      `json:"height,omitempty"`
	Thumbnail               string                   `json:"thumbnail,omitempty"`
	Caption                 string                   `json:"caption,omitempty"`
	ParseMode               string                   `json:"parse_mode,omitempty"`
	CaptionEntities         []MessageEntity          `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia   bool                     `json:"show_caption_above_media,omitempty"`
	HasSpoiler              bool                     `json:"has_spoiler,omitempty"`
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

func (p SendAnimationParams) base() SendBaseParams {
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
func (p SendAnimationParams) caption() CaptionParams {
	return CaptionParams{
		Caption:               p.Caption,
		ParseMode:             p.ParseMode,
		CaptionEntities:       p.CaptionEntities,
		ShowCaptionAboveMedia: p.ShowCaptionAboveMedia,
	}
}
func (b *Client) SendAnimation(ctx context.Context, params SendAnimationParams) (*Message, error) {
	return b.sendMediaJSON(ctx, "sendAnimation", params.ChatID, params.Animation, params)
}
func (b *Client) SendAnimationUpload(ctx context.Context, params SendAnimationParams, name string, reader io.Reader) (*Message, error) {
	fields, err := mediaFields(params.base(), params.caption())
	if err != nil {
		return nil, err
	}
	fields.Int("duration", params.Duration)
	fields.Int("width", params.Width)
	fields.Int("height", params.Height)
	fields.String("thumbnail", params.Thumbnail)
	fields.Bool("has_spoiler", params.HasSpoiler)
	return b.sendUpload(ctx, "sendAnimation", "animation", fields, name, reader)
}
