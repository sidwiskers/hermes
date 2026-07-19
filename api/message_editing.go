package api

import (
	"context"
	"fmt"
	"strings"
)

// InputEditableMedia is media accepted by editMessageMedia. It is closed to
// Telegram's supported animation, audio, document, live-photo, photo, and
// video variants.
type InputEditableMedia interface {
	inputEditableMedia()
	editableMediaSource() string
}

func (InputMedia) inputEditableMedia()               {}
func (media InputMedia) editableMediaSource() string { return media.Media }

func (InputMediaAnimation) inputEditableMedia()               {}
func (media InputMediaAnimation) editableMediaSource() string { return media.Media }

func (InputMediaAudio) inputEditableMedia()               {}
func (media InputMediaAudio) editableMediaSource() string { return media.Media }

func (InputMediaDocument) inputEditableMedia()               {}
func (media InputMediaDocument) editableMediaSource() string { return media.Media }

func (InputMediaLivePhoto) inputEditableMedia()               {}
func (media InputMediaLivePhoto) editableMediaSource() string { return media.Media }

func (InputMediaPhoto) inputEditableMedia()               {}
func (media InputMediaPhoto) editableMediaSource() string { return media.Media }

func (InputMediaVideo) inputEditableMedia()               {}
func (media InputMediaVideo) editableMediaSource() string { return media.Media }

func validateMessageEditTarget(chatID any, messageID int, inlineMessageID, method string) error {
	if strings.TrimSpace(inlineMessageID) != "" {
		if chatID != nil || messageID != 0 {
			return fmt.Errorf("hermes: %s inline_message_id cannot be combined with chat_id or message_id", method)
		}
		return nil
	}
	if err := validateChatID(chatID, method); err != nil {
		return err
	}
	if messageID == 0 {
		return fmt.Errorf("hermes: %s message_id is required", method)
	}
	return nil
}

type EditMessageMediaParams struct {
	BusinessConnectionID string                `json:"business_connection_id,omitempty"`
	ChatID               any                   `json:"chat_id,omitempty"`
	MessageID            int                   `json:"message_id,omitempty"`
	InlineMessageID      string                `json:"inline_message_id,omitempty"`
	Media                InputEditableMedia    `json:"media"`
	ReplyMarkup          *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func validateEditMessageMedia(params EditMessageMediaParams) error {
	if err := validateMessageEditTarget(params.ChatID, params.MessageID, params.InlineMessageID, "editMessageMedia"); err != nil {
		return err
	}
	if params.Media == nil || strings.TrimSpace(params.Media.editableMediaSource()) == "" {
		return fmt.Errorf("hermes: editMessageMedia media is required")
	}
	if livePhoto, ok := params.Media.(InputMediaLivePhoto); ok && strings.TrimSpace(livePhoto.Photo) == "" {
		return fmt.Errorf("hermes: editMessageMedia live photo requires photo")
	}
	return nil
}

func (client *Client) EditMessageMedia(ctx context.Context, params EditMessageMediaParams) (*Message, error) {
	if err := validateEditMessageMedia(params); err != nil {
		return nil, err
	}
	if err := validateAttachmentUploads(params.Media, nil, "editMessageMedia"); err != nil {
		return nil, err
	}
	return callMessageOrBool(ctx, client, "editMessageMedia", params)
}

// EditMessageMediaUpload edits a non-inline message and streams every
// attach:// reference found in Media.
func (client *Client) EditMessageMediaUpload(ctx context.Context, params EditMessageMediaParams, uploads ...Upload) (*Message, error) {
	if err := validateEditMessageMedia(params); err != nil {
		return nil, err
	}
	if strings.TrimSpace(params.InlineMessageID) != "" {
		return nil, fmt.Errorf("hermes: editMessageMedia inline messages cannot upload files")
	}
	if len(uploads) == 0 {
		return client.EditMessageMedia(ctx, params)
	}
	if err := validateAttachmentUploads(params.Media, uploads, "editMessageMedia"); err != nil {
		return nil, err
	}
	fields, err := newFormFields(params.ChatID)
	if err != nil {
		return nil, err
	}
	fields.String("business_connection_id", params.BusinessConnectionID)
	fields.Int("message_id", params.MessageID)
	if err = fields.JSON("media", params.Media); err != nil {
		return nil, err
	}
	if params.ReplyMarkup != nil {
		if err = fields.JSON("reply_markup", params.ReplyMarkup); err != nil {
			return nil, err
		}
	}
	var message Message
	if err = client.CallMultipart(ctx, "editMessageMedia", fields, uploads, &message); err != nil {
		return nil, err
	}
	return &message, nil
}

type EditMessageLiveLocationParams struct {
	BusinessConnectionID string                `json:"business_connection_id,omitempty"`
	ChatID               any                   `json:"chat_id,omitempty"`
	MessageID            int                   `json:"message_id,omitempty"`
	InlineMessageID      string                `json:"inline_message_id,omitempty"`
	Latitude             float64               `json:"latitude"`
	Longitude            float64               `json:"longitude"`
	LivePeriod           int                   `json:"live_period,omitempty"`
	HorizontalAccuracy   float64               `json:"horizontal_accuracy,omitempty"`
	Heading              int                   `json:"heading,omitempty"`
	ProximityAlertRadius int                   `json:"proximity_alert_radius,omitempty"`
	ReplyMarkup          *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (client *Client) EditMessageLiveLocation(ctx context.Context, params EditMessageLiveLocationParams) (*Message, error) {
	if err := validateMessageEditTarget(params.ChatID, params.MessageID, params.InlineMessageID, "editMessageLiveLocation"); err != nil {
		return nil, err
	}
	if params.Latitude < -90 || params.Latitude > 90 || params.Longitude < -180 || params.Longitude > 180 {
		return nil, fmt.Errorf("hermes: editMessageLiveLocation coordinates are out of range")
	}
	if params.HorizontalAccuracy < 0 || params.HorizontalAccuracy > 1500 {
		return nil, fmt.Errorf("hermes: editMessageLiveLocation horizontal_accuracy must be between 0 and 1500")
	}
	if params.Heading < 0 || params.Heading > 360 {
		return nil, fmt.Errorf("hermes: editMessageLiveLocation heading must be between 1 and 360 when set")
	}
	if params.ProximityAlertRadius < 0 || params.ProximityAlertRadius > 100000 {
		return nil, fmt.Errorf("hermes: editMessageLiveLocation proximity_alert_radius must be between 1 and 100000 when set")
	}
	return callMessageOrBool(ctx, client, "editMessageLiveLocation", params)
}

type StopMessageLiveLocationParams struct {
	BusinessConnectionID string                `json:"business_connection_id,omitempty"`
	ChatID               any                   `json:"chat_id,omitempty"`
	MessageID            int                   `json:"message_id,omitempty"`
	InlineMessageID      string                `json:"inline_message_id,omitempty"`
	ReplyMarkup          *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (client *Client) StopMessageLiveLocation(ctx context.Context, params StopMessageLiveLocationParams) (*Message, error) {
	if err := validateMessageEditTarget(params.ChatID, params.MessageID, params.InlineMessageID, "stopMessageLiveLocation"); err != nil {
		return nil, err
	}
	return callMessageOrBool(ctx, client, "stopMessageLiveLocation", params)
}
