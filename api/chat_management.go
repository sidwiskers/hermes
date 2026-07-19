package api

import (
	"context"
	"fmt"
	"io"
	"strings"
)

func (b *Client) SetChatPhoto(ctx context.Context, chatID any, filename string, reader io.Reader) error {
	if err := validateChatID(chatID, "setChatPhoto"); err != nil {
		return err
	}
	if reader == nil {
		return fmt.Errorf("hermes: setChatPhoto upload reader is required")
	}
	fields, err := newFormFields(chatID)
	if err != nil {
		return err
	}
	var ok bool
	if err = b.CallMultipart(ctx, "setChatPhoto", fields, []Upload{{Field: "photo", Name: filename, Reader: reader}}, &ok); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("hermes: setChatPhoto returned false")
	}
	return nil
}

func (b *Client) DeleteChatPhoto(ctx context.Context, chatID any) error {
	if err := validateChatID(chatID, "deleteChatPhoto"); err != nil {
		return err
	}
	return b.callTrue(ctx, "deleteChatPhoto", GetChatParams{ChatID: chatID})
}

type SetChatTitleParams struct {
	ChatID any    `json:"chat_id"`
	Title  string `json:"title"`
}

func (b *Client) SetChatTitle(ctx context.Context, params SetChatTitleParams) error {
	if err := validateChatID(params.ChatID, "setChatTitle"); err != nil {
		return err
	}
	length := len([]rune(strings.TrimSpace(params.Title)))
	if length < 1 || length > 128 {
		return fmt.Errorf("hermes: chat title must contain 1-128 characters")
	}
	return b.callTrue(ctx, "setChatTitle", params)
}

type SetChatDescriptionParams struct {
	ChatID      any    `json:"chat_id"`
	Description string `json:"description,omitempty"`
}

func (b *Client) SetChatDescription(ctx context.Context, params SetChatDescriptionParams) error {
	if err := validateChatID(params.ChatID, "setChatDescription"); err != nil {
		return err
	}
	if len([]rune(params.Description)) > 255 {
		return fmt.Errorf("hermes: chat description must not exceed 255 characters")
	}
	return b.callTrue(ctx, "setChatDescription", params)
}

type PinChatMessageParams struct {
	BusinessConnectionID string `json:"business_connection_id,omitempty"`
	ChatID               any    `json:"chat_id"`
	MessageID            int    `json:"message_id"`
	DisableNotification  bool   `json:"disable_notification,omitempty"`
}

func (b *Client) PinChatMessage(ctx context.Context, params PinChatMessageParams) error {
	if err := validateMessageTarget(params.ChatID, params.MessageID, "pinChatMessage"); err != nil {
		return err
	}
	return b.callTrue(ctx, "pinChatMessage", params)
}

type UnpinChatMessageParams struct {
	BusinessConnectionID string `json:"business_connection_id,omitempty"`
	ChatID               any    `json:"chat_id"`
	MessageID            int    `json:"message_id,omitempty"`
}

func (b *Client) UnpinChatMessage(ctx context.Context, params UnpinChatMessageParams) error {
	if err := validateChatID(params.ChatID, "unpinChatMessage"); err != nil {
		return err
	}
	if params.BusinessConnectionID != "" && params.MessageID == 0 {
		return fmt.Errorf("hermes: unpinChatMessage message_id is required for business messages")
	}
	return b.callTrue(ctx, "unpinChatMessage", params)
}

func (b *Client) UnpinAllChatMessages(ctx context.Context, chatID any) error {
	if err := validateChatID(chatID, "unpinAllChatMessages"); err != nil {
		return err
	}
	return b.callTrue(ctx, "unpinAllChatMessages", GetChatParams{ChatID: chatID})
}

type SetChatStickerSetParams struct {
	ChatID         any    `json:"chat_id"`
	StickerSetName string `json:"sticker_set_name"`
}

func (b *Client) SetChatStickerSet(ctx context.Context, params SetChatStickerSetParams) error {
	if err := validateChatID(params.ChatID, "setChatStickerSet"); err != nil {
		return err
	}
	if strings.TrimSpace(params.StickerSetName) == "" {
		return fmt.Errorf("hermes: setChatStickerSet sticker_set_name is required")
	}
	return b.callTrue(ctx, "setChatStickerSet", params)
}

func (b *Client) DeleteChatStickerSet(ctx context.Context, chatID any) error {
	if err := validateChatID(chatID, "deleteChatStickerSet"); err != nil {
		return err
	}
	return b.callTrue(ctx, "deleteChatStickerSet", GetChatParams{ChatID: chatID})
}

type GetUserPersonalChatMessagesParams struct {
	UserID int64 `json:"user_id"`
	Limit  int   `json:"limit"`
}

func (b *Client) GetUserPersonalChatMessages(ctx context.Context, params GetUserPersonalChatMessagesParams) ([]Message, error) {
	if params.UserID == 0 {
		return nil, fmt.Errorf("hermes: getUserPersonalChatMessages user_id is required")
	}
	if params.Limit < 1 || params.Limit > 20 {
		return nil, fmt.Errorf("hermes: getUserPersonalChatMessages limit must be 1-20")
	}
	var messages []Message
	if err := b.Call(ctx, "getUserPersonalChatMessages", params, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}
