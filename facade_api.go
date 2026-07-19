package hermes

import (
	"context"
	"io"

	"github.com/sidwiskers/hermes/api"
	runtimecore "github.com/sidwiskers/hermes/internal/runtime"
)

// Low-level API request and error types are re-exported from the root facade.
type (
	APIClient      = api.Client
	APIError       = api.APIError
	HTTPError      = api.HTTPError
	TransportError = api.TransportError
	APIObserver    = api.Observer
	APICallKind    = api.CallKind
	APICallEvent   = api.CallEvent
	APICallResult  = api.CallResult
	Upload         = api.Upload

	SendBaseParams               = api.SendBaseParams
	SendMessageParams            = api.SendMessageParams
	EditMessageTextParams        = api.EditMessageTextParams
	EditMessageCaptionParams     = api.EditMessageCaptionParams
	EditMessageReplyMarkupParams = api.EditMessageReplyMarkupParams
	DeleteMessageParams          = api.DeleteMessageParams
	DeleteMessagesParams         = api.DeleteMessagesParams
	ForwardMessageParams         = api.ForwardMessageParams
	ForwardMessagesParams        = api.ForwardMessagesParams
	CopyMessageParams            = api.CopyMessageParams
	CopyMessagesParams           = api.CopyMessagesParams
	MessageID                    = api.MessageID
	SendMessageDraftParams       = api.SendMessageDraftParams
	AnswerCallbackQueryParams    = api.AnswerCallbackQueryParams
	SendChatActionParams         = api.SendChatActionParams

	CaptionParams                           = api.CaptionParams
	SendPhotoParams                         = api.SendPhotoParams
	SendLivePhotoParams                     = api.SendLivePhotoParams
	SendAnimationParams                     = api.SendAnimationParams
	SendAudioParams                         = api.SendAudioParams
	SendDocumentParams                      = api.SendDocumentParams
	SendStickerParams                       = api.SendStickerParams
	SendVideoParams                         = api.SendVideoParams
	SendVideoNoteParams                     = api.SendVideoNoteParams
	SendVoiceParams                         = api.SendVoiceParams
	SendContactParams                       = api.SendContactParams
	SendLocationParams                      = api.SendLocationParams
	SendVenueParams                         = api.SendVenueParams
	RichMessageMedia                        = api.RichMessageMedia
	InputMediaVoiceNote                     = api.InputMediaVoiceNote
	InputRichMessageMedia                   = api.InputRichMessageMedia
	InputRichMessage                        = api.InputRichMessage
	InputRichMessageContent                 = api.InputRichMessageContent
	InputRichBlock                          = api.InputRichBlock
	InputRichBlockParagraph                 = api.InputRichBlockParagraph
	InputRichBlockSectionHeading            = api.InputRichBlockSectionHeading
	InputRichBlockPreformatted              = api.InputRichBlockPreformatted
	InputRichBlockFooter                    = api.InputRichBlockFooter
	InputRichBlockDivider                   = api.InputRichBlockDivider
	InputRichBlockMathematicalExpression    = api.InputRichBlockMathematicalExpression
	InputRichBlockAnchor                    = api.InputRichBlockAnchor
	InputRichBlockListItem                  = api.InputRichBlockListItem
	InputRichBlockList                      = api.InputRichBlockList
	InputRichBlockBlockQuotation            = api.InputRichBlockBlockQuotation
	InputRichBlockPullQuotation             = api.InputRichBlockPullQuotation
	InputRichBlockCollage                   = api.InputRichBlockCollage
	InputRichBlockSlideshow                 = api.InputRichBlockSlideshow
	InputRichBlockTable                     = api.InputRichBlockTable
	InputRichBlockDetails                   = api.InputRichBlockDetails
	InputRichBlockMap                       = api.InputRichBlockMap
	InputRichBlockAnimation                 = api.InputRichBlockAnimation
	InputRichBlockAudio                     = api.InputRichBlockAudio
	InputRichBlockPhoto                     = api.InputRichBlockPhoto
	InputRichBlockVideo                     = api.InputRichBlockVideo
	InputRichBlockVoiceNote                 = api.InputRichBlockVoiceNote
	InputRichBlockThinking                  = api.InputRichBlockThinking
	SendRichMessageParams                   = api.SendRichMessageParams
	SendRichMessageDraftParams              = api.SendRichMessageDraftParams
	SendInvoiceParams                       = api.SendInvoiceParams
	CreateInvoiceLinkParams                 = api.CreateInvoiceLinkParams
	AnswerShippingQueryParams               = api.AnswerShippingQueryParams
	AnswerPreCheckoutQueryParams            = api.AnswerPreCheckoutQueryParams
	GetStarTransactionsParams               = api.GetStarTransactionsParams
	RefundStarPaymentParams                 = api.RefundStarPaymentParams
	EditUserStarSubscriptionParams          = api.EditUserStarSubscriptionParams
	InputPaidMedia                          = api.InputPaidMedia
	InputPaidMediaLivePhoto                 = api.InputPaidMediaLivePhoto
	InputPaidMediaPhoto                     = api.InputPaidMediaPhoto
	InputPaidMediaVideo                     = api.InputPaidMediaVideo
	SendPaidMediaParams                     = api.SendPaidMediaParams
	CreateForumTopicParams                  = api.CreateForumTopicParams
	EditForumTopicParams                    = api.EditForumTopicParams
	ForumTopicTargetParams                  = api.ForumTopicTargetParams
	EditGeneralForumTopicParams             = api.EditGeneralForumTopicParams
	GeneralForumTopicParams                 = api.GeneralForumTopicParams
	InputEditableMedia                      = api.InputEditableMedia
	EditMessageMediaParams                  = api.EditMessageMediaParams
	EditMessageLiveLocationParams           = api.EditMessageLiveLocationParams
	StopMessageLiveLocationParams           = api.StopMessageLiveLocationParams
	InputSticker                            = api.InputSticker
	GetStickerSetParams                     = api.GetStickerSetParams
	GetCustomEmojiStickersParams            = api.GetCustomEmojiStickersParams
	UploadStickerFileParams                 = api.UploadStickerFileParams
	CreateNewStickerSetParams               = api.CreateNewStickerSetParams
	AddStickerToSetParams                   = api.AddStickerToSetParams
	ReplaceStickerInSetParams               = api.ReplaceStickerInSetParams
	StickerFileParams                       = api.StickerFileParams
	SetStickerPositionInSetParams           = api.SetStickerPositionInSetParams
	SetStickerEmojiListParams               = api.SetStickerEmojiListParams
	SetStickerKeywordsParams                = api.SetStickerKeywordsParams
	SetStickerMaskPositionParams            = api.SetStickerMaskPositionParams
	SetStickerSetTitleParams                = api.SetStickerSetTitleParams
	SetStickerSetThumbnailParams            = api.SetStickerSetThumbnailParams
	SetCustomEmojiStickerSetThumbnailParams = api.SetCustomEmojiStickerSetThumbnailParams
	DeleteStickerSetParams                  = api.DeleteStickerSetParams
	InlineQueryResult                       = api.InlineQueryResult
	InlineQueryResultRaw                    = api.InlineQueryResultRaw
	AnswerInlineQueryParams                 = api.AnswerInlineQueryParams
	AnswerWebAppQueryParams                 = api.AnswerWebAppQueryParams
	SavePreparedInlineMessageParams         = api.SavePreparedInlineMessageParams
	SavePreparedKeyboardButtonParams        = api.SavePreparedKeyboardButtonParams
	AnswerGuestQueryParams                  = api.AnswerGuestQueryParams
	SendGameParams                          = api.SendGameParams
	SetGameScoreParams                      = api.SetGameScoreParams
	GetGameHighScoresParams                 = api.GetGameHighScoresParams
	PassportElementError                    = api.PassportElementError
	PassportElementErrorRaw                 = api.PassportElementErrorRaw
	SetPassportDataErrorsParams             = api.SetPassportDataErrorsParams
	ApproveSuggestedPostParams              = api.ApproveSuggestedPostParams
	DeclineSuggestedPostParams              = api.DeclineSuggestedPostParams
	ManagedBotParams                        = api.ManagedBotParams
	SetManagedBotAccessSettingsParams       = api.SetManagedBotAccessSettingsParams
	InputProfilePhoto                       = api.InputProfilePhoto
	InputProfilePhotoStatic                 = api.InputProfilePhotoStatic
	InputProfilePhotoAnimated               = api.InputProfilePhotoAnimated
	SetMyProfilePhotoParams                 = api.SetMyProfilePhotoParams
	ReadBusinessMessageParams               = api.ReadBusinessMessageParams
	DeleteBusinessMessagesParams            = api.DeleteBusinessMessagesParams
	SetBusinessAccountNameParams            = api.SetBusinessAccountNameParams
	SetBusinessAccountUsernameParams        = api.SetBusinessAccountUsernameParams
	SetBusinessAccountBioParams             = api.SetBusinessAccountBioParams
	SetBusinessAccountProfilePhotoParams    = api.SetBusinessAccountProfilePhotoParams
	RemoveBusinessAccountProfilePhotoParams = api.RemoveBusinessAccountProfilePhotoParams
	SetBusinessAccountGiftSettingsParams    = api.SetBusinessAccountGiftSettingsParams
	TransferBusinessAccountStarsParams      = api.TransferBusinessAccountStarsParams
	GetBusinessAccountStarBalanceParams     = api.GetBusinessAccountStarBalanceParams
	SendGiftParams                          = api.SendGiftParams
	GiftPremiumSubscriptionParams           = api.GiftPremiumSubscriptionParams
	OwnedGiftsFilter                        = api.OwnedGiftsFilter
	GetBusinessAccountGiftsParams           = api.GetBusinessAccountGiftsParams
	GetUserGiftsParams                      = api.GetUserGiftsParams
	GetChatGiftsParams                      = api.GetChatGiftsParams
	OwnedGiftParams                         = api.OwnedGiftParams
	UpgradeGiftParams                       = api.UpgradeGiftParams
	TransferGiftParams                      = api.TransferGiftParams
	InputStoryContent                       = api.InputStoryContent
	InputStoryContentPhoto                  = api.InputStoryContentPhoto
	InputStoryContentVideo                  = api.InputStoryContentVideo
	StoryAreaPosition                       = api.StoryAreaPosition
	StoryAreaType                           = api.StoryAreaType
	StoryArea                               = api.StoryArea
	PostStoryParams                         = api.PostStoryParams
	RepostStoryParams                       = api.RepostStoryParams
	EditStoryParams                         = api.EditStoryParams
	DeleteStoryParams                       = api.DeleteStoryParams
	InputChecklistTask                      = api.InputChecklistTask
	InputChecklist                          = api.InputChecklist
	SendChecklistParams                     = api.SendChecklistParams
	EditMessageChecklistParams              = api.EditMessageChecklistParams
	GetUserProfilePhotosParams              = api.GetUserProfilePhotosParams
	GetUserProfileAudiosParams              = api.GetUserProfileAudiosParams
	SetUserEmojiStatusParams                = api.SetUserEmojiStatusParams
	GetUserChatBoostsParams                 = api.GetUserChatBoostsParams
	GetBusinessConnectionParams             = api.GetBusinessConnectionParams
	SetMyNameParams                         = api.SetMyNameParams
	GetMyNameParams                         = api.GetMyNameParams
	SetMyDescriptionParams                  = api.SetMyDescriptionParams
	GetMyDescriptionParams                  = api.GetMyDescriptionParams
	SetMyShortDescriptionParams             = api.SetMyShortDescriptionParams
	GetMyShortDescriptionParams             = api.GetMyShortDescriptionParams
	SetChatMenuButtonParams                 = api.SetChatMenuButtonParams
	GetChatMenuButtonParams                 = api.GetChatMenuButtonParams
	SetMyDefaultAdministratorRightsParams   = api.SetMyDefaultAdministratorRightsParams
	GetMyDefaultAdministratorRightsParams   = api.GetMyDefaultAdministratorRightsParams
	VerifyUserParams                        = api.VerifyUserParams
	VerifyChatParams                        = api.VerifyChatParams
	RemoveUserVerificationParams            = api.RemoveUserVerificationParams
	RemoveChatVerificationParams            = api.RemoveChatVerificationParams

	EditEphemeralMessageTextParams        = api.EditEphemeralMessageTextParams
	InputMedia                            = api.InputMedia
	EditEphemeralMessageMediaParams       = api.EditEphemeralMessageMediaParams
	EditEphemeralMessageCaptionParams     = api.EditEphemeralMessageCaptionParams
	EditEphemeralMessageReplyMarkupParams = api.EditEphemeralMessageReplyMarkupParams
	DeleteEphemeralMessageParams          = api.DeleteEphemeralMessageParams

	MediaGroupItem                  = api.MediaGroupItem
	InputPollMedia                  = api.InputPollMedia
	InputPollOptionMedia            = api.InputPollOptionMedia
	InputMediaPhoto                 = api.InputMediaPhoto
	InputMediaVideo                 = api.InputMediaVideo
	InputMediaAudio                 = api.InputMediaAudio
	InputMediaDocument              = api.InputMediaDocument
	InputMediaAnimation             = api.InputMediaAnimation
	InputMediaLivePhoto             = api.InputMediaLivePhoto
	InputMediaLocation              = api.InputMediaLocation
	InputMediaVenue                 = api.InputMediaVenue
	InputMediaLink                  = api.InputMediaLink
	InputMediaSticker               = api.InputMediaSticker
	SendMediaGroupParams            = api.SendMediaGroupParams
	InputPollOption                 = api.InputPollOption
	SendPollParams                  = api.SendPollParams
	StopPollParams                  = api.StopPollParams
	SendDiceParams                  = api.SendDiceParams
	SetMessageReactionParams        = api.SetMessageReactionParams
	DeleteMessageReactionParams     = api.DeleteMessageReactionParams
	DeleteAllMessageReactionsParams = api.DeleteAllMessageReactionsParams

	BanChatMemberParams                   = api.BanChatMemberParams
	UnbanChatMemberParams                 = api.UnbanChatMemberParams
	RestrictChatMemberParams              = api.RestrictChatMemberParams
	PromoteChatMemberParams               = api.PromoteChatMemberParams
	SetChatAdministratorCustomTitleParams = api.SetChatAdministratorCustomTitleParams
	SetChatMemberTagParams                = api.SetChatMemberTagParams
	ChatSenderParams                      = api.ChatSenderParams
	SetChatPermissionsParams              = api.SetChatPermissionsParams
	GetChatAdministratorsParams           = api.GetChatAdministratorsParams
	GetChatMemberParams                   = api.GetChatMemberParams
	ChatJoinRequestParams                 = api.ChatJoinRequestParams
	AnswerChatJoinRequestQueryParams      = api.AnswerChatJoinRequestQueryParams
	SendChatJoinRequestWebAppParams       = api.SendChatJoinRequestWebAppParams

	ExportChatInviteLinkParams             = api.ExportChatInviteLinkParams
	CreateChatInviteLinkParams             = api.CreateChatInviteLinkParams
	EditChatInviteLinkParams               = api.EditChatInviteLinkParams
	CreateChatSubscriptionInviteLinkParams = api.CreateChatSubscriptionInviteLinkParams
	EditChatSubscriptionInviteLinkParams   = api.EditChatSubscriptionInviteLinkParams
	RevokeChatInviteLinkParams             = api.RevokeChatInviteLinkParams

	SetChatTitleParams                = api.SetChatTitleParams
	SetChatDescriptionParams          = api.SetChatDescriptionParams
	PinChatMessageParams              = api.PinChatMessageParams
	UnpinChatMessageParams            = api.UnpinChatMessageParams
	SetChatStickerSetParams           = api.SetChatStickerSetParams
	GetUserPersonalChatMessagesParams = api.GetUserPersonalChatMessagesParams

	GetFileParams                        = api.GetFileParams
	GetUpdatesParams                     = api.GetUpdatesParams
	SetWebhookParams                     = api.SetWebhookParams
	DeleteWebhookParams                  = api.DeleteWebhookParams
	WebhookInfo                          = api.WebhookInfo
	GetChatParams                        = api.GetChatParams
	BotCommandScope                      = api.BotCommandScope
	BotCommandScopeDefault               = api.BotCommandScopeDefault
	BotCommandScopeAllPrivateChats       = api.BotCommandScopeAllPrivateChats
	BotCommandScopeAllGroupChats         = api.BotCommandScopeAllGroupChats
	BotCommandScopeAllChatAdministrators = api.BotCommandScopeAllChatAdministrators
	BotCommandScopeChat                  = api.BotCommandScopeChat
	BotCommandScopeChatAdministrators    = api.BotCommandScopeChatAdministrators
	BotCommandScopeChatMember            = api.BotCommandScopeChatMember
	SetMyCommandsParams                  = api.SetMyCommandsParams
	DeleteMyCommandsParams               = api.DeleteMyCommandsParams
	GetMyCommandsParams                  = api.GetMyCommandsParams
)

var (
	ErrClientRequired         = api.ErrClientRequired
	ErrTokenRequired          = api.ErrTokenRequired
	ErrInvalidMethod          = api.ErrInvalidMethod
	ErrResponseTooLarge       = api.ErrResponseTooLarge
	ErrResultMissing          = api.ErrResultMissing
	ErrQueueFull              = runtimecore.ErrQueueFull
	ErrWebhookHandlerRequired = runtimecore.ErrWebhookHandlerRequired
	ErrWebhookReplyInvalid    = runtimecore.ErrWebhookReplyInvalid
	ErrWebhookReplyTooLarge   = runtimecore.ErrWebhookReplyTooLarge
	ErrUpdateSourceRequired   = runtimecore.ErrUpdateSourceRequired
	ErrDispatchRequired       = runtimecore.ErrDispatchRequired
)

const (
	ParseModeHTML         = api.ParseModeHTML
	ParseModeMarkdown     = api.ParseModeMarkdown
	ParseModeMarkdownV2   = api.ParseModeMarkdownV2
	ActionTyping          = api.ActionTyping
	ActionUploadPhoto     = api.ActionUploadPhoto
	ActionRecordVideo     = api.ActionRecordVideo
	ActionUploadVideo     = api.ActionUploadVideo
	ActionRecordVoice     = api.ActionRecordVoice
	ActionUploadVoice     = api.ActionUploadVoice
	ActionUploadDocument  = api.ActionUploadDocument
	ActionChooseSticker   = api.ActionChooseSticker
	ActionFindLocation    = api.ActionFindLocation
	ActionRecordVideoNote = api.ActionRecordVideoNote
	ActionUploadVideoNote = api.ActionUploadVideoNote

	PollRegular     = api.PollRegular
	PollQuiz        = api.PollQuiz
	DiceEmoji       = api.DiceEmoji
	DartsEmoji      = api.DartsEmoji
	BasketballEmoji = api.BasketballEmoji
	FootballEmoji   = api.FootballEmoji
	BowlingEmoji    = api.BowlingEmoji
	SlotsEmoji      = api.SlotsEmoji

	JoinRequestApprove     = api.JoinRequestApprove
	JoinRequestDecline     = api.JoinRequestDecline
	JoinRequestQueue       = api.JoinRequestQueue
	ForumIconBlue          = api.ForumIconBlue
	ForumIconYellow        = api.ForumIconYellow
	ForumIconPurple        = api.ForumIconPurple
	ForumIconGreen         = api.ForumIconGreen
	ForumIconPink          = api.ForumIconPink
	ForumIconRed           = api.ForumIconRed
	MenuButtonTypeCommands = api.MenuButtonTypeCommands
	MenuButtonTypeWebApp   = api.MenuButtonTypeWebApp
	MenuButtonTypeDefault  = api.MenuButtonTypeDefault
	StickerFormatStatic    = api.StickerFormatStatic
	StickerFormatAnimated  = api.StickerFormatAnimated
	StickerFormatVideo     = api.StickerFormatVideo
	StickerTypeRegular     = api.StickerTypeRegular
	StickerTypeMask        = api.StickerTypeMask
	StickerTypeCustomEmoji = api.StickerTypeCustomEmoji
	StoryActive6Hours      = api.StoryActive6Hours
	StoryActive12Hours     = api.StoryActive12Hours
	StoryActive24Hours     = api.StoryActive24Hours
	StoryActive48Hours     = api.StoryActive48Hours
)

// Call invokes any Bot API method and decodes its result as T. It is the
// forward-compatible escape hatch for methods not yet present in the typed
// client.
func Call[T any](ctx context.Context, bot *Bot, method string, params any) (T, error) {
	var zero T
	if bot == nil || bot.Client == nil {
		return zero, api.ErrClientRequired
	}
	return api.Call[T](ctx, bot.Client, method, params)
}

// MultipartJSON encodes a structured multipart field as JSON.
func MultipartJSON(value any) (string, error) { return api.MultipartJSON(value) }

// MultipartInt formats an integer multipart field.
func MultipartInt(value int64) string { return api.MultipartInt(value) }

// Bool returns a pointer that preserves an explicitly supplied false value.
func Bool(value bool) *bool { return api.Bool(value) }

// Attachment returns Telegram's attach:// reference for a multipart field.
func Attachment(field string) string { return api.Attachment(field) }

// NewUpload describes one streamed multipart upload. The caller retains
// ownership of reader.
func NewUpload(field, filename string, reader io.Reader) Upload {
	return api.NewUpload(field, filename, reader)
}

func DefaultCommandScope() BotCommandScopeDefault { return api.DefaultCommandScope() }
func AllPrivateChatsCommandScope() BotCommandScopeAllPrivateChats {
	return api.AllPrivateChatsCommandScope()
}
func AllGroupChatsCommandScope() BotCommandScopeAllGroupChats {
	return api.AllGroupChatsCommandScope()
}
func AllChatAdministratorsCommandScope() BotCommandScopeAllChatAdministrators {
	return api.AllChatAdministratorsCommandScope()
}
func ChatCommandScope(chatID any) BotCommandScopeChat { return api.ChatCommandScope(chatID) }
func ChatAdministratorsCommandScope(chatID any) BotCommandScopeChatAdministrators {
	return api.ChatAdministratorsCommandScope(chatID)
}
func ChatMemberCommandScope(chatID any, userID int64) BotCommandScopeChatMember {
	return api.ChatMemberCommandScope(chatID, userID)
}

func CommandsMenuButton() MenuButton               { return api.CommandsMenuButton() }
func WebAppMenuButton(text, url string) MenuButton { return api.WebAppMenuButton(text, url) }
func DefaultMenuButton() MenuButton                { return api.DefaultMenuButton() }
