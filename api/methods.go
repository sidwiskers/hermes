package api

import (
	"context"
	"fmt"
	"strings"
)

const (
	ParseModeHTML       = "HTML"
	ParseModeMarkdown   = "Markdown"
	ParseModeMarkdownV2 = "MarkdownV2"
)

// SendBaseParams contains options shared by most send methods.
// ReceiverUserID and CallbackQueryID activate Bot API 10.2 ephemeral delivery
// on methods that support it.
type SendBaseParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
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

type SendMessageParams struct {
	BusinessConnectionID    string                   `json:"business_connection_id,omitempty"`
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	Text                    string                   `json:"text"`
	ParseMode               string                   `json:"parse_mode,omitempty"`
	Entities                []MessageEntity          `json:"entities,omitempty"`
	LinkPreviewOptions      *LinkPreviewOptions      `json:"link_preview_options,omitempty"`
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

func (b *Client) SendMessage(ctx context.Context, params SendMessageParams) (*Message, error) {
	if err := validateChatID(params.ChatID, "sendMessage"); err != nil {
		return nil, err
	}
	if params.Text == "" {
		return nil, fmt.Errorf("hermes: sendMessage text is required")
	}
	return callMessage(ctx, b, "sendMessage", params)
}

type EditMessageTextParams struct {
	BusinessConnectionID string                `json:"business_connection_id,omitempty"`
	ChatID               any                   `json:"chat_id,omitempty"`
	MessageID            int                   `json:"message_id,omitempty"`
	InlineMessageID      string                `json:"inline_message_id,omitempty"`
	Text                 string                `json:"text,omitempty"`
	ParseMode            string                `json:"parse_mode,omitempty"`
	Entities             []MessageEntity       `json:"entities,omitempty"`
	LinkPreviewOptions   *LinkPreviewOptions   `json:"link_preview_options,omitempty"`
	ReplyMarkup          *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	RichMessage          *InputRichMessage     `json:"rich_message,omitempty"`
}

func (b *Client) EditMessageText(ctx context.Context, params EditMessageTextParams) (*Message, error) {
	if params.Text == "" && params.RichMessage == nil {
		return nil, fmt.Errorf("hermes: editMessageText text or rich_message is required")
	}
	if params.RichMessage != nil {
		if err := validateRichMessage(*params.RichMessage, false); err != nil {
			return nil, err
		}
		if err := validateAttachmentUploads(*params.RichMessage, nil, "editMessageText"); err != nil {
			return nil, err
		}
	}
	return callMessageOrBool(ctx, b, "editMessageText", params)
}

type EditMessageCaptionParams struct {
	BusinessConnectionID  string                `json:"business_connection_id,omitempty"`
	ChatID                any                   `json:"chat_id,omitempty"`
	MessageID             int                   `json:"message_id,omitempty"`
	InlineMessageID       string                `json:"inline_message_id,omitempty"`
	Caption               string                `json:"caption,omitempty"`
	ParseMode             string                `json:"parse_mode,omitempty"`
	CaptionEntities       []MessageEntity       `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia bool                  `json:"show_caption_above_media,omitempty"`
	ReplyMarkup           *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (b *Client) EditMessageCaption(ctx context.Context, params EditMessageCaptionParams) (*Message, error) {
	return callMessageOrBool(ctx, b, "editMessageCaption", params)
}

type EditMessageReplyMarkupParams struct {
	BusinessConnectionID string                `json:"business_connection_id,omitempty"`
	ChatID               any                   `json:"chat_id,omitempty"`
	MessageID            int                   `json:"message_id,omitempty"`
	InlineMessageID      string                `json:"inline_message_id,omitempty"`
	ReplyMarkup          *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (b *Client) EditMessageReplyMarkup(ctx context.Context, params EditMessageReplyMarkupParams) (*Message, error) {
	return callMessageOrBool(ctx, b, "editMessageReplyMarkup", params)
}

type DeleteMessageParams struct {
	ChatID    any `json:"chat_id"`
	MessageID int `json:"message_id"`
}

func (b *Client) DeleteMessage(ctx context.Context, params DeleteMessageParams) error {
	if err := validateChatID(params.ChatID, "deleteMessage"); err != nil {
		return err
	}
	if params.MessageID == 0 {
		return fmt.Errorf("hermes: deleteMessage message_id is required")
	}
	return b.callTrue(ctx, "deleteMessage", params)
}

type DeleteMessagesParams struct {
	ChatID     any   `json:"chat_id"`
	MessageIDs []int `json:"message_ids"`
}

func (b *Client) DeleteMessages(ctx context.Context, params DeleteMessagesParams) error {
	if err := validateChatID(params.ChatID, "deleteMessages"); err != nil {
		return err
	}
	if len(params.MessageIDs) == 0 || len(params.MessageIDs) > 100 {
		return fmt.Errorf("hermes: deleteMessages requires 1-100 message_ids")
	}
	return b.callTrue(ctx, "deleteMessages", params)
}

type ForwardMessageParams struct {
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	FromChatID              any                      `json:"from_chat_id"`
	VideoStartTimestamp     int                      `json:"video_start_timestamp,omitempty"`
	DisableNotification     bool                     `json:"disable_notification,omitempty"`
	ProtectContent          bool                     `json:"protect_content,omitempty"`
	MessageEffectID         string                   `json:"message_effect_id,omitempty"`
	SuggestedPostParameters *SuggestedPostParameters `json:"suggested_post_parameters,omitempty"`
	MessageID               int                      `json:"message_id"`
}

func (b *Client) ForwardMessage(ctx context.Context, params ForwardMessageParams) (*Message, error) {
	if err := validateChatID(params.ChatID, "forwardMessage"); err != nil {
		return nil, err
	}
	if err := validateChatID(params.FromChatID, "forwardMessage"); err != nil || params.MessageID == 0 {
		return nil, fmt.Errorf("hermes: forwardMessage from_chat_id and message_id are required")
	}
	return callMessage(ctx, b, "forwardMessage", params)
}

type CopyMessageParams struct {
	ChatID                  any                      `json:"chat_id"`
	MessageThreadID         int                      `json:"message_thread_id,omitempty"`
	DirectMessagesTopicID   int                      `json:"direct_messages_topic_id,omitempty"`
	FromChatID              any                      `json:"from_chat_id"`
	MessageID               int                      `json:"message_id"`
	VideoStartTimestamp     int                      `json:"video_start_timestamp,omitempty"`
	Caption                 string                   `json:"caption,omitempty"`
	ParseMode               string                   `json:"parse_mode,omitempty"`
	CaptionEntities         []MessageEntity          `json:"caption_entities,omitempty"`
	ShowCaptionAboveMedia   bool                     `json:"show_caption_above_media,omitempty"`
	DisableNotification     bool                     `json:"disable_notification,omitempty"`
	ProtectContent          bool                     `json:"protect_content,omitempty"`
	AllowPaidBroadcast      bool                     `json:"allow_paid_broadcast,omitempty"`
	MessageEffectID         string                   `json:"message_effect_id,omitempty"`
	SuggestedPostParameters *SuggestedPostParameters `json:"suggested_post_parameters,omitempty"`
	ReplyParameters         *ReplyParameters         `json:"reply_parameters,omitempty"`
	ReplyMarkup             ReplyMarkup              `json:"reply_markup,omitempty"`
}

type MessageID struct {
	MessageID int `json:"message_id"`
}

func (b *Client) CopyMessage(ctx context.Context, params CopyMessageParams) (int, error) {
	if err := validateChatID(params.ChatID, "copyMessage"); err != nil {
		return 0, err
	}
	if err := validateChatID(params.FromChatID, "copyMessage"); err != nil || params.MessageID == 0 {
		return 0, fmt.Errorf("hermes: copyMessage from_chat_id and message_id are required")
	}
	var result MessageID
	if err := b.Call(ctx, "copyMessage", params, &result); err != nil {
		return 0, err
	}
	return result.MessageID, nil
}

type AnswerCallbackQueryParams struct {
	CallbackQueryID string `json:"callback_query_id"`
	Text            string `json:"text,omitempty"`
	ShowAlert       bool   `json:"show_alert,omitempty"`
	URL             string `json:"url,omitempty"`
	CacheTime       int    `json:"cache_time,omitempty"`
}

func (b *Client) AnswerCallback(ctx context.Context, params AnswerCallbackQueryParams) error {
	if strings.TrimSpace(params.CallbackQueryID) == "" {
		return fmt.Errorf("hermes: callback_query_id is required")
	}
	return b.callTrue(ctx, "answerCallbackQuery", params)
}

const (
	ActionTyping          = "typing"
	ActionUploadPhoto     = "upload_photo"
	ActionRecordVideo     = "record_video"
	ActionUploadVideo     = "upload_video"
	ActionRecordVoice     = "record_voice"
	ActionUploadVoice     = "upload_voice"
	ActionUploadDocument  = "upload_document"
	ActionChooseSticker   = "choose_sticker"
	ActionFindLocation    = "find_location"
	ActionRecordVideoNote = "record_video_note"
	ActionUploadVideoNote = "upload_video_note"
)

type SendChatActionParams struct {
	BusinessConnectionID string `json:"business_connection_id,omitempty"`
	ChatID               any    `json:"chat_id"`
	MessageThreadID      int    `json:"message_thread_id,omitempty"`
	Action               string `json:"action"`
}

func (b *Client) SendChatAction(ctx context.Context, params SendChatActionParams) error {
	if err := validateChatID(params.ChatID, "sendChatAction"); err != nil {
		return err
	}
	if params.Action == "" {
		return fmt.Errorf("hermes: sendChatAction action is required")
	}
	return b.callTrue(ctx, "sendChatAction", params)
}

func (b *Client) GetMe(ctx context.Context) (*User, error) {
	var user User
	if err := b.Call(ctx, "getMe", nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}
