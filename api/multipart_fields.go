package api

import (
	"fmt"
	"strconv"
)

type formFields map[string]string

func newFormFields(chatID any) (formFields, error) {
	fields := make(formFields, 24)
	if chatID == nil {
		return nil, fmt.Errorf("hermes: chat_id is required")
	}
	switch value := chatID.(type) {
	case int:
		fields["chat_id"] = strconv.Itoa(value)
	case int32:
		fields["chat_id"] = strconv.FormatInt(int64(value), 10)
	case int64:
		fields["chat_id"] = strconv.FormatInt(value, 10)
	case string:
		fields["chat_id"] = value
	default:
		encoded, err := MultipartJSON(value)
		if err != nil {
			return nil, fmt.Errorf("hermes: encode chat_id: %w", err)
		}
		fields["chat_id"] = encoded
	}
	return fields, nil
}

func (f formFields) String(key, value string) {
	if value != "" {
		f[key] = value
	}
}

func (f formFields) Bool(key string, value bool) {
	if value {
		f[key] = "true"
	}
}

func (f formFields) BoolPointer(key string, value *bool) {
	if value != nil {
		f[key] = strconv.FormatBool(*value)
	}
}

func (f formFields) Int(key string, value int) {
	if value != 0 {
		f[key] = strconv.Itoa(value)
	}
}

func (f formFields) Int64(key string, value int64) {
	if value != 0 {
		f[key] = strconv.FormatInt(value, 10)
	}
}

func (f formFields) Float(key string, value float64) {
	if value != 0 {
		f[key] = strconv.FormatFloat(value, 'f', -1, 64)
	}
}

func (f formFields) JSON(key string, value any) error {
	if value == nil {
		return nil
	}
	encoded, err := MultipartJSON(value)
	if err != nil {
		return fmt.Errorf("hermes: encode multipart field %s: %w", key, err)
	}
	f[key] = encoded
	return nil
}

func addSendBaseFields(fields formFields, params SendBaseParams) error {
	fields.String("business_connection_id", params.BusinessConnectionID)
	fields.Int("message_thread_id", params.MessageThreadID)
	fields.Int("direct_messages_topic_id", params.DirectMessagesTopicID)
	fields.Bool("disable_notification", params.DisableNotification)
	fields.Bool("protect_content", params.ProtectContent)
	fields.Bool("allow_paid_broadcast", params.AllowPaidBroadcast)
	fields.String("message_effect_id", params.MessageEffectID)
	if params.SuggestedPostParameters != nil {
		if err := fields.JSON("suggested_post_parameters", params.SuggestedPostParameters); err != nil {
			return err
		}
	}
	fields.Int64("receiver_user_id", params.ReceiverUserID)
	fields.String("callback_query_id", params.CallbackQueryID)
	if params.ReplyParameters != nil {
		if err := fields.JSON("reply_parameters", params.ReplyParameters); err != nil {
			return err
		}
	}
	if params.ReplyMarkup != nil {
		if err := fields.JSON("reply_markup", params.ReplyMarkup); err != nil {
			return err
		}
	}
	return nil
}

func addCaptionFields(fields formFields, params CaptionParams) error {
	fields.String("caption", params.Caption)
	fields.String("parse_mode", params.ParseMode)
	fields.Bool("show_caption_above_media", params.ShowCaptionAboveMedia)
	if len(params.CaptionEntities) != 0 {
		return fields.JSON("caption_entities", params.CaptionEntities)
	}
	return nil
}
