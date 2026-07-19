package api

import (
	"context"
	"fmt"
	"strings"
)

type SendLivePhotoParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	ReceiverUserID          int64                    `json:"receiver_user_id,omitempty"`
	CallbackQueryID         string                   `json:"callback_query_id,omitempty"`
	LivePhoto               string                   `json:"live_photo"`
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
}

func validateSendLivePhoto(params SendLivePhotoParams) error {
	if err := validateChatID(params.ChatID, "sendLivePhoto"); err != nil {
		return err
	}
	if strings.TrimSpace(params.LivePhoto) == "" || strings.TrimSpace(params.Photo) == "" {
		return fmt.Errorf("hermes: sendLivePhoto live_photo and photo are required")
	}
	return nil
}

func (client *Client) SendLivePhoto(ctx context.Context, params SendLivePhotoParams) (*Message, error) {
	if err := validateSendLivePhoto(params); err != nil {
		return nil, err
	}
	if err := validateAttachmentUploads(params, nil, "sendLivePhoto"); err != nil {
		return nil, err
	}
	return callMessage(ctx, client, "sendLivePhoto", params)
}

func (client *Client) SendLivePhotoUpload(ctx context.Context, params SendLivePhotoParams, uploads ...Upload) (*Message, error) {
	if err := validateSendLivePhoto(params); err != nil {
		return nil, err
	}
	if len(uploads) == 0 {
		return client.SendLivePhoto(ctx, params)
	}
	if err := validateAttachmentUploads(params, uploads, "sendLivePhoto"); err != nil {
		return nil, err
	}
	fields, err := newFormFields(params.ChatID)
	if err != nil {
		return nil, err
	}
	fields.String("business_connection_id", params.BusinessConnectionID)
	fields.Int("message_thread_id", params.MessageThreadID)
	fields.Int("direct_messages_topic_id", params.DirectMessagesTopicID)
	fields.Int64("receiver_user_id", params.ReceiverUserID)
	fields.String("callback_query_id", params.CallbackQueryID)
	fields.String("live_photo", params.LivePhoto)
	fields.String("photo", params.Photo)
	fields.String("caption", params.Caption)
	fields.String("parse_mode", params.ParseMode)
	if len(params.CaptionEntities) != 0 {
		if err = fields.JSON("caption_entities", params.CaptionEntities); err != nil {
			return nil, err
		}
	}
	fields.Bool("show_caption_above_media", params.ShowCaptionAboveMedia)
	fields.Bool("has_spoiler", params.HasSpoiler)
	fields.Bool("disable_notification", params.DisableNotification)
	fields.Bool("protect_content", params.ProtectContent)
	fields.Bool("allow_paid_broadcast", params.AllowPaidBroadcast)
	fields.String("message_effect_id", params.MessageEffectID)
	if params.SuggestedPostParameters != nil {
		if err = fields.JSON("suggested_post_parameters", params.SuggestedPostParameters); err != nil {
			return nil, err
		}
	}
	if params.ReplyParameters != nil {
		if err = fields.JSON("reply_parameters", params.ReplyParameters); err != nil {
			return nil, err
		}
	}
	if params.ReplyMarkup != nil {
		if err = fields.JSON("reply_markup", params.ReplyMarkup); err != nil {
			return nil, err
		}
	}
	var message Message
	if err = client.CallMultipart(ctx, "sendLivePhoto", fields, uploads, &message); err != nil {
		return nil, err
	}
	return &message, nil
}
