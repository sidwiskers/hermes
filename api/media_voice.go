package api

import (
	"context"
	"io"
)

type SendVoiceParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	Voice                   string                   `json:"voice"`
	Caption                 string                   `json:"caption,omitempty"`
	ParseMode               string                   `json:"parse_mode,omitempty"`
	CaptionEntities         []MessageEntity          `json:"caption_entities,omitempty"`
	Duration                int                      `json:"duration,omitempty"`
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

func (p SendVoiceParams) base() SendBaseParams {
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
func (p SendVoiceParams) caption() CaptionParams {
	return CaptionParams{Caption: p.Caption, ParseMode: p.ParseMode, CaptionEntities: p.CaptionEntities}
}
func (b *Client) SendVoice(ctx context.Context, p SendVoiceParams) (*Message, error) {
	return b.sendMediaJSON(ctx, "sendVoice", p.ChatID, p.Voice, p)
}
func (b *Client) SendVoiceUpload(ctx context.Context, p SendVoiceParams, name string, reader io.Reader) (*Message, error) {
	fields, err := mediaFields(p.base(), p.caption())
	if err != nil {
		return nil, err
	}
	fields.Int("duration", p.Duration)
	return b.sendUpload(ctx, "sendVoice", "voice", fields, name, reader)
}
