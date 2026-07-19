package api

import (
	"context"
	"io"
)

type SendAudioParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	Audio                   string                   `json:"audio"`
	Caption                 string                   `json:"caption,omitempty"`
	ParseMode               string                   `json:"parse_mode,omitempty"`
	CaptionEntities         []MessageEntity          `json:"caption_entities,omitempty"`
	Duration                int                      `json:"duration,omitempty"`
	Performer               string                   `json:"performer,omitempty"`
	Title                   string                   `json:"title,omitempty"`
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

func (p SendAudioParams) base() SendBaseParams {
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
func (p SendAudioParams) caption() CaptionParams {
	return CaptionParams{Caption: p.Caption, ParseMode: p.ParseMode, CaptionEntities: p.CaptionEntities}
}
func (b *Client) SendAudio(ctx context.Context, params SendAudioParams) (*Message, error) {
	return b.sendMediaJSON(ctx, "sendAudio", params.ChatID, params.Audio, params)
}
func (b *Client) SendAudioUpload(ctx context.Context, params SendAudioParams, name string, reader io.Reader) (*Message, error) {
	fields, err := mediaFields(params.base(), params.caption())
	if err != nil {
		return nil, err
	}
	fields.Int("duration", params.Duration)
	fields.String("performer", params.Performer)
	fields.String("title", params.Title)
	fields.String("thumbnail", params.Thumbnail)
	return b.sendUpload(ctx, "sendAudio", "audio", fields, name, reader)
}
