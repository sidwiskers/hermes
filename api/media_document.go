package api

import (
	"context"
	"io"
)

type SendDocumentParams struct {
	BusinessConnectionID        string                   `json:"business_connection_id,omitempty"`
	ChatID                      any                      `json:"chat_id"`
	MessageThreadID             int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID       int                      `json:"direct_messages_topic_id,omitempty"`
	Document                    string                   `json:"document"`
	Thumbnail                   string                   `json:"thumbnail,omitempty"`
	Caption                     string                   `json:"caption,omitempty"`
	ParseMode                   string                   `json:"parse_mode,omitempty"`
	CaptionEntities             []MessageEntity          `json:"caption_entities,omitempty"`
	DisableContentTypeDetection bool                     `json:"disable_content_type_detection,omitempty"`
	DisableNotification         bool                     `json:"disable_notification,omitempty"`
	ProtectContent              bool                     `json:"protect_content,omitempty"`
	AllowPaidBroadcast          bool                     `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID             string                   `json:"message_effect_id,omitempty"`
	SuggestedPostParameters     *SuggestedPostParameters `json:"suggested_post_parameters,omitempty"`
	ReplyParameters             *ReplyParameters         `json:"reply_parameters,omitempty"`
	ReplyMarkup                 ReplyMarkup              `json:"reply_markup,omitempty"`
	ReceiverUserID              int64                    `json:"receiver_user_id,omitempty"`
	CallbackQueryID             string                   `json:"callback_query_id,omitempty"`
}

func (p SendDocumentParams) base() SendBaseParams {
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
func (p SendDocumentParams) caption() CaptionParams {
	return CaptionParams{Caption: p.Caption, ParseMode: p.ParseMode, CaptionEntities: p.CaptionEntities}
}
func (b *Client) SendDocument(ctx context.Context, p SendDocumentParams) (*Message, error) {
	return b.sendMediaJSON(ctx, "sendDocument", p.ChatID, p.Document, p)
}
func (b *Client) SendDocumentUpload(ctx context.Context, p SendDocumentParams, name string, reader io.Reader) (*Message, error) {
	fields, err := mediaFields(p.base(), p.caption())
	if err != nil {
		return nil, err
	}
	fields.String("thumbnail", p.Thumbnail)
	fields.Bool("disable_content_type_detection", p.DisableContentTypeDetection)
	return b.sendUpload(ctx, "sendDocument", "document", fields, name, reader)
}

type SendStickerParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	Sticker                 string                   `json:"sticker"`
	Emoji                   string                   `json:"emoji,omitempty"`
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

func (p SendStickerParams) base() SendBaseParams {
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
func (b *Client) SendSticker(ctx context.Context, p SendStickerParams) (*Message, error) {
	return b.sendMediaJSON(ctx, "sendSticker", p.ChatID, p.Sticker, p)
}
func (b *Client) SendStickerUpload(ctx context.Context, p SendStickerParams, name string, reader io.Reader) (*Message, error) {
	fields, err := newFormFields(p.ChatID)
	if err != nil {
		return nil, err
	}
	if err = addSendBaseFields(fields, p.base()); err != nil {
		return nil, err
	}
	fields.String("emoji", p.Emoji)
	return b.sendUpload(ctx, "sendSticker", "sticker", fields, name, reader)
}
