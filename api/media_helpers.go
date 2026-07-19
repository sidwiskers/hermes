package api

import (
	"context"
	"fmt"
	"io"
)

// CaptionParams contains fields shared by captioned media methods.
type CaptionParams struct {
	Caption               string          `json:"caption,omitempty"`
	ParseMode             string          `json:"parse_mode,omitempty"`
	CaptionEntities       []MessageEntity `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia bool            `json:"show_caption_above_media,omitempty"`
}

func mediaFields(base SendBaseParams, caption CaptionParams) (formFields, error) {
	fields, err := newFormFields(base.ChatID)
	if err != nil {
		return nil, err
	}
	if err = addSendBaseFields(fields, base); err != nil {
		return nil, err
	}
	if err = addCaptionFields(fields, caption); err != nil {
		return nil, err
	}
	return fields, nil
}
func (b *Client) sendMediaJSON(ctx context.Context, method string, chatID any, media string, params any) (*Message, error) {
	if err := validateChatID(chatID, method); err != nil {
		return nil, err
	}
	if media == "" {
		return nil, fmt.Errorf("hermes: %s media is required", method)
	}
	return callMessage(ctx, b, method, params)
}
func (b *Client) sendUpload(ctx context.Context, method, field string, fields formFields, name string, reader io.Reader) (*Message, error) {
	if reader == nil {
		return nil, fmt.Errorf("hermes: %s upload reader is required", method)
	}
	fields[field] = "attach://" + field
	var message Message
	if err := b.CallMultipart(ctx, method, fields, []Upload{{Field: field, Name: name, Reader: reader}}, &message); err != nil {
		return nil, err
	}
	return &message, nil
}
