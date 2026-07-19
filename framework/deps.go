package framework

import (
	"github.com/sidwiskers/hermes/api"
	telegram "github.com/sidwiskers/hermes/types"
)

// Telegram schemas and Bot API parameters used by the framework helpers are
// aliases of their canonical types and remain assignment-compatible.
type (
	Update               = telegram.Update
	UpdateType           = telegram.UpdateType
	User                 = telegram.User
	Chat                 = telegram.Chat
	Message              = telegram.Message
	CallbackQuery        = telegram.CallbackQuery
	ReplyParameters      = telegram.ReplyParameters
	LinkPreviewOptions   = telegram.LinkPreviewOptions
	ReplyMarkup          = telegram.ReplyMarkup
	InlineKeyboardMarkup = telegram.InlineKeyboardMarkup
	EphemeralMessageRef  = telegram.EphemeralMessageRef
	ReactionType         = telegram.ReactionType
	ChatPermissions      = telegram.ChatPermissions

	SendMessageParams                     = api.SendMessageParams
	InputRichMessage                      = api.InputRichMessage
	SendRichMessageParams                 = api.SendRichMessageParams
	AnswerCallbackQueryParams             = api.AnswerCallbackQueryParams
	SendChatActionParams                  = api.SendChatActionParams
	EditMessageTextParams                 = api.EditMessageTextParams
	EditMessageCaptionParams              = api.EditMessageCaptionParams
	EditMessageReplyMarkupParams          = api.EditMessageReplyMarkupParams
	DeleteMessageParams                   = api.DeleteMessageParams
	EditEphemeralMessageTextParams        = api.EditEphemeralMessageTextParams
	EditEphemeralMessageCaptionParams     = api.EditEphemeralMessageCaptionParams
	EditEphemeralMessageReplyMarkupParams = api.EditEphemeralMessageReplyMarkupParams
	DeleteEphemeralMessageParams          = api.DeleteEphemeralMessageParams
	SendPhotoParams                       = api.SendPhotoParams
	SendDocumentParams                    = api.SendDocumentParams
	SendVideoParams                       = api.SendVideoParams
	SendAnimationParams                   = api.SendAnimationParams
	SendAudioParams                       = api.SendAudioParams
	SendVoiceParams                       = api.SendVoiceParams
	SendStickerParams                     = api.SendStickerParams
	SetMessageReactionParams              = api.SetMessageReactionParams
	SendDiceParams                        = api.SendDiceParams
	InputPollOption                       = api.InputPollOption
	SendPollParams                        = api.SendPollParams
	BanChatMemberParams                   = api.BanChatMemberParams
	ChatJoinRequestParams                 = api.ChatJoinRequestParams
	PinChatMessageParams                  = api.PinChatMessageParams
	UnpinChatMessageParams                = api.UnpinChatMessageParams
)

// Common update and parse-mode constants used by Context helpers.
const (
	UpdateUnknown       = telegram.UpdateUnknown
	ParseModeHTML       = api.ParseModeHTML
	ParseModeMarkdown   = api.ParseModeMarkdown
	ParseModeMarkdownV2 = api.ParseModeMarkdownV2
)
