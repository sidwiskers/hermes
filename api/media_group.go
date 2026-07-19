package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// MediaGroupItem is one item in a Telegram album.
type MediaGroupItem interface {
	mediaGroupItem()
	mediaGroupType() string
	mediaGroupSource() string
}

// Attachment returns the attach:// reference used by multipart Bot API calls.
func Attachment(field string) string { return "attach://" + strings.TrimSpace(field) }

// NewUpload creates a streamed multipart upload.
func NewUpload(field, filename string, reader io.Reader) Upload {
	return Upload{Field: field, Name: filename, Reader: reader}
}

type InputMediaPhoto struct {
	Media                 string          `json:"media"`
	Caption               string          `json:"caption,omitempty"`
	ParseMode             string          `json:"parse_mode,omitempty"`
	CaptionEntities       []MessageEntity `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia bool            `json:"show_caption_above_media,omitempty"`
	HasSpoiler            bool            `json:"has_spoiler,omitempty"`
}

func (InputMediaPhoto) mediaGroupItem()            {}
func (InputMediaPhoto) mediaGroupType() string     { return "photo" }
func (m InputMediaPhoto) mediaGroupSource() string { return m.Media }
func (m InputMediaPhoto) MarshalJSON() ([]byte, error) {
	type alias InputMediaPhoto
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "photo", alias: alias(m)})
}

type InputMediaVideo struct {
	Media                 string          `json:"media"`
	Thumbnail             string          `json:"thumbnail,omitempty"`
	Cover                 string          `json:"cover,omitempty"`
	StartTimestamp        int             `json:"start_timestamp,omitempty"`
	Caption               string          `json:"caption,omitempty"`
	ParseMode             string          `json:"parse_mode,omitempty"`
	CaptionEntities       []MessageEntity `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia bool            `json:"show_caption_above_media,omitempty"`
	Width                 int             `json:"width,omitempty"`
	Height                int             `json:"height,omitempty"`
	Duration              int             `json:"duration,omitempty"`
	SupportsStreaming     bool            `json:"supports_streaming,omitempty"`
	HasSpoiler            bool            `json:"has_spoiler,omitempty"`
}

func (InputMediaVideo) mediaGroupItem()            {}
func (InputMediaVideo) mediaGroupType() string     { return "video" }
func (m InputMediaVideo) mediaGroupSource() string { return m.Media }
func (m InputMediaVideo) MarshalJSON() ([]byte, error) {
	type alias InputMediaVideo
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "video", alias: alias(m)})
}

type InputMediaAudio struct {
	Media           string          `json:"media"`
	Thumbnail       string          `json:"thumbnail,omitempty"`
	Caption         string          `json:"caption,omitempty"`
	ParseMode       string          `json:"parse_mode,omitempty"`
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"`
	Duration        int             `json:"duration,omitempty"`
	Performer       string          `json:"performer,omitempty"`
	Title           string          `json:"title,omitempty"`
}

func (InputMediaAudio) mediaGroupItem()            {}
func (InputMediaAudio) mediaGroupType() string     { return "audio" }
func (m InputMediaAudio) mediaGroupSource() string { return m.Media }
func (m InputMediaAudio) MarshalJSON() ([]byte, error) {
	type alias InputMediaAudio
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "audio", alias: alias(m)})
}

type InputMediaDocument struct {
	Media                       string          `json:"media"`
	Thumbnail                   string          `json:"thumbnail,omitempty"`
	Caption                     string          `json:"caption,omitempty"`
	ParseMode                   string          `json:"parse_mode,omitempty"`
	CaptionEntities             []MessageEntity `json:"caption_entities,omitempty"`
	DisableContentTypeDetection bool            `json:"disable_content_type_detection,omitempty"`
}

func (InputMediaDocument) mediaGroupItem()            {}
func (InputMediaDocument) mediaGroupType() string     { return "document" }
func (m InputMediaDocument) mediaGroupSource() string { return m.Media }
func (m InputMediaDocument) MarshalJSON() ([]byte, error) {
	type alias InputMediaDocument
	return json.Marshal(struct {
		Type string `json:"type"`
		alias
	}{Type: "document", alias: alias(m)})
}

type SendMediaGroupParams struct {
	BusinessConnectionID  string           `json:"business_connection_id,omitempty"`
	ChatID                any              `json:"chat_id"`
	MessageThreadID       int              `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID int              `json:"direct_messages_topic_id,omitempty"`
	Media                 []MediaGroupItem `json:"media"`
	DisableNotification   bool             `json:"disable_notification,omitempty"`
	ProtectContent        bool             `json:"protect_content,omitempty"`
	AllowPaidBroadcast    bool             `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID       string           `json:"message_effect_id,omitempty"`
	ReplyParameters       *ReplyParameters `json:"reply_parameters,omitempty"`
}

func validateMediaGroup(params SendMediaGroupParams) error {
	if err := validateChatID(params.ChatID, "sendMediaGroup"); err != nil {
		return err
	}
	if len(params.Media) < 2 || len(params.Media) > 10 {
		return fmt.Errorf("hermes: sendMediaGroup requires 2-10 items")
	}
	kind := ""
	for index, item := range params.Media {
		if item == nil || strings.TrimSpace(item.mediaGroupSource()) == "" {
			return fmt.Errorf("hermes: sendMediaGroup item %d has no media", index)
		}
		if live, ok := item.(InputMediaLivePhoto); ok && strings.TrimSpace(live.Photo) == "" {
			return fmt.Errorf("hermes: sendMediaGroup live-photo item %d has no static photo", index)
		}
		current := item.mediaGroupType()
		if current == "audio" || current == "document" {
			if kind == "" {
				kind = current
			}
			if kind != current {
				return fmt.Errorf("hermes: audio and document albums must contain one media type")
			}
		} else if kind == "audio" || kind == "document" {
			return fmt.Errorf("hermes: audio and document albums must contain one media type")
		}
	}
	return nil
}

func (b *Client) SendMediaGroup(ctx context.Context, params SendMediaGroupParams) ([]Message, error) {
	if err := validateMediaGroup(params); err != nil {
		return nil, err
	}
	var messages []Message
	if err := b.Call(ctx, "sendMediaGroup", params, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

// SendMediaGroupUpload sends an album with one or more streamed attachments.
// Album items refer to each upload with Attachment(upload.Field).
func (b *Client) SendMediaGroupUpload(ctx context.Context, params SendMediaGroupParams, uploads ...Upload) ([]Message, error) {
	if err := validateMediaGroup(params); err != nil {
		return nil, err
	}
	if len(uploads) == 0 {
		return b.SendMediaGroup(ctx, params)
	}
	fields, err := newFormFields(params.ChatID)
	if err != nil {
		return nil, err
	}
	fields.String("business_connection_id", params.BusinessConnectionID)
	fields.Int("message_thread_id", params.MessageThreadID)
	fields.Int("direct_messages_topic_id", params.DirectMessagesTopicID)
	fields.Bool("disable_notification", params.DisableNotification)
	fields.Bool("protect_content", params.ProtectContent)
	fields.Bool("allow_paid_broadcast", params.AllowPaidBroadcast)
	fields.String("message_effect_id", params.MessageEffectID)
	if err = fields.JSON("media", params.Media); err != nil {
		return nil, err
	}
	if params.ReplyParameters != nil {
		if err = fields.JSON("reply_parameters", params.ReplyParameters); err != nil {
			return nil, err
		}
	}

	if err = validateAttachmentUploads(params.Media, uploads, "sendMediaGroup"); err != nil {
		return nil, err
	}

	var messages []Message
	if err = b.CallMultipart(ctx, "sendMediaGroup", fields, uploads, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}
