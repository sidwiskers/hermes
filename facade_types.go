package hermes

import telegram "github.com/sidwiskers/hermes/types"

// Core Telegram Bot API types are re-exported from the root package for the
// ergonomic hermes.X surface. The canonical schema package is types.
type (
	ResponseParameters             = telegram.ResponseParameters
	UpdateType                     = telegram.UpdateType
	Update                         = telegram.Update
	User                           = telegram.User
	Chat                           = telegram.Chat
	ChatFullInfo                   = telegram.ChatFullInfo
	Message                        = telegram.Message
	CallbackQuery                  = telegram.CallbackQuery
	MessageEntity                  = telegram.MessageEntity
	ReplyParameters                = telegram.ReplyParameters
	LinkPreviewOptions             = telegram.LinkPreviewOptions
	EphemeralMessageRef            = telegram.EphemeralMessageRef
	BotCommand                     = telegram.BotCommand
	PhotoSize                      = telegram.PhotoSize
	File                           = telegram.File
	Animation                      = telegram.Animation
	Audio                          = telegram.Audio
	Document                       = telegram.Document
	LivePhoto                      = telegram.LivePhoto
	Sticker                        = telegram.Sticker
	StickerSet                     = telegram.StickerSet
	MaskPosition                   = telegram.MaskPosition
	Video                          = telegram.Video
	VideoNote                      = telegram.VideoNote
	Voice                          = telegram.Voice
	Contact                        = telegram.Contact
	Location                       = telegram.Location
	Venue                          = telegram.Venue
	Dice                           = telegram.Dice
	Link                           = telegram.Link
	PollMedia                      = telegram.PollMedia
	PollOption                     = telegram.PollOption
	Poll                           = telegram.Poll
	PollAnswer                     = telegram.PollAnswer
	InlineQuery                    = telegram.InlineQuery
	ChosenInlineResult             = telegram.ChosenInlineResult
	ShippingAddress                = telegram.ShippingAddress
	ShippingQuery                  = telegram.ShippingQuery
	OrderInfo                      = telegram.OrderInfo
	PreCheckoutQuery               = telegram.PreCheckoutQuery
	BusinessConnection             = telegram.BusinessConnection
	BusinessMessagesDeleted        = telegram.BusinessMessagesDeleted
	ChatMemberUpdated              = telegram.ChatMemberUpdated
	ChatJoinRequest                = telegram.ChatJoinRequest
	MessageReactionUpdated         = telegram.MessageReactionUpdated
	MessageReactionCountUpdated    = telegram.MessageReactionCountUpdated
	Community                      = telegram.Community
	CommunityChatAdded             = telegram.CommunityChatAdded
	CommunityChatRemoved           = telegram.CommunityChatRemoved
	BotSubscriptionUpdated         = telegram.BotSubscriptionUpdated
	ChatPermissions                = telegram.ChatPermissions
	ChatMember                     = telegram.ChatMember
	ChatInviteLink                 = telegram.ChatInviteLink
	ReactionType                   = telegram.ReactionType
	ReactionCount                  = telegram.ReactionCount
	RichText                       = telegram.RichText
	RichTextBold                   = telegram.RichTextBold
	RichTextItalic                 = telegram.RichTextItalic
	RichTextUnderline              = telegram.RichTextUnderline
	RichTextStrikethrough          = telegram.RichTextStrikethrough
	RichTextSpoiler                = telegram.RichTextSpoiler
	RichTextDateTime               = telegram.RichTextDateTime
	RichTextTextMention            = telegram.RichTextTextMention
	RichTextSubscript              = telegram.RichTextSubscript
	RichTextSuperscript            = telegram.RichTextSuperscript
	RichTextMarked                 = telegram.RichTextMarked
	RichTextCode                   = telegram.RichTextCode
	RichTextCustomEmoji            = telegram.RichTextCustomEmoji
	RichTextMathematicalExpression = telegram.RichTextMathematicalExpression
	RichTextURL                    = telegram.RichTextURL
	RichTextEmailAddress           = telegram.RichTextEmailAddress
	RichTextPhoneNumber            = telegram.RichTextPhoneNumber
	RichTextBankCardNumber         = telegram.RichTextBankCardNumber
	RichTextMention                = telegram.RichTextMention
	RichTextHashtag                = telegram.RichTextHashtag
	RichTextCashtag                = telegram.RichTextCashtag
	RichTextBotCommand             = telegram.RichTextBotCommand
	RichTextAnchor                 = telegram.RichTextAnchor
	RichTextAnchorLink             = telegram.RichTextAnchorLink
	RichTextReference              = telegram.RichTextReference
	RichTextReferenceLink          = telegram.RichTextReferenceLink
	RichBlockCaption               = telegram.RichBlockCaption
	RichBlockTableCell             = telegram.RichBlockTableCell
	RichBlockListItem              = telegram.RichBlockListItem
	RichBlock                      = telegram.RichBlock
	RichMessage                    = telegram.RichMessage
	LabeledPrice                   = telegram.LabeledPrice
	Invoice                        = telegram.Invoice
	ShippingOption                 = telegram.ShippingOption
	SuccessfulPayment              = telegram.SuccessfulPayment
	RefundedPayment                = telegram.RefundedPayment
	PaidMedia                      = telegram.PaidMedia
	PaidMediaInfo                  = telegram.PaidMediaInfo
	PaidMediaPurchased             = telegram.PaidMediaPurchased
	StarAmount                     = telegram.StarAmount
	RevenueWithdrawalState         = telegram.RevenueWithdrawalState
	GiftBackground                 = telegram.GiftBackground
	Gift                           = telegram.Gift
	AffiliateInfo                  = telegram.AffiliateInfo
	TransactionPartner             = telegram.TransactionPartner
	StarTransaction                = telegram.StarTransaction
	StarTransactions               = telegram.StarTransactions
	ForumTopic                     = telegram.ForumTopic
	BotName                        = telegram.BotName
	BotDescription                 = telegram.BotDescription
	BotShortDescription            = telegram.BotShortDescription
	UserProfilePhotos              = telegram.UserProfilePhotos
	UserProfileAudios              = telegram.UserProfileAudios
	ChatAdministratorRights        = telegram.ChatAdministratorRights
	MenuButton                     = telegram.MenuButton
	BusinessBotRights              = telegram.BusinessBotRights
	ChatBoostSource                = telegram.ChatBoostSource
	ChatBoost                      = telegram.ChatBoost
	UserChatBoosts                 = telegram.UserChatBoosts
	ChatBoostUpdated               = telegram.ChatBoostUpdated
	ChatBoostRemoved               = telegram.ChatBoostRemoved
	InlineQueryResultsButton       = telegram.InlineQueryResultsButton
	SentWebAppMessage              = telegram.SentWebAppMessage
	PreparedInlineMessage          = telegram.PreparedInlineMessage
	PreparedKeyboardButton         = telegram.PreparedKeyboardButton
	SentGuestMessage               = telegram.SentGuestMessage
	Game                           = telegram.Game
	GameHighScore                  = telegram.GameHighScore
	BotAccessSettings              = telegram.BotAccessSettings
	ManagedBotCreated              = telegram.ManagedBotCreated
	ManagedBotUpdated              = telegram.ManagedBotUpdated
	AcceptedGiftTypes              = telegram.AcceptedGiftTypes
	Gifts                          = telegram.Gifts
	UniqueGift                     = telegram.UniqueGift
	OwnedGift                      = telegram.OwnedGift
	OwnedGifts                     = telegram.OwnedGifts
	Story                          = telegram.Story
	ChecklistTask                  = telegram.ChecklistTask
	Checklist                      = telegram.Checklist

	ReplyMarkup                     = telegram.ReplyMarkup
	InlineKeyboardMarkup            = telegram.InlineKeyboardMarkup
	InlineKeyboardButton            = telegram.InlineKeyboardButton
	WebAppInfo                      = telegram.WebAppInfo
	LoginURL                        = telegram.LoginURL
	SwitchInlineQueryChosenChat     = telegram.SwitchInlineQueryChosenChat
	CopyTextButton                  = telegram.CopyTextButton
	CallbackGame                    = telegram.CallbackGame
	ReplyKeyboardMarkup             = telegram.ReplyKeyboardMarkup
	KeyboardButton                  = telegram.KeyboardButton
	KeyboardButtonRequestUsers      = telegram.KeyboardButtonRequestUsers
	KeyboardButtonRequestChat       = telegram.KeyboardButtonRequestChat
	KeyboardButtonRequestManagedBot = telegram.KeyboardButtonRequestManagedBot
	KeyboardButtonPollType          = telegram.KeyboardButtonPollType
	ReplyKeyboardRemove             = telegram.ReplyKeyboardRemove
	ForceReply                      = telegram.ForceReply
	ForceReplyMarkup                = telegram.ForceReplyMarkup
)

const (
	ReactionEmoji       = telegram.ReactionEmoji
	ReactionCustomEmoji = telegram.ReactionCustomEmoji
	ReactionPaid        = telegram.ReactionPaid

	UpdateUnknown                 = telegram.UpdateUnknown
	UpdateMessage                 = telegram.UpdateMessage
	UpdateEditedMessage           = telegram.UpdateEditedMessage
	UpdateChannelPost             = telegram.UpdateChannelPost
	UpdateEditedChannelPost       = telegram.UpdateEditedChannelPost
	UpdateBusinessConnection      = telegram.UpdateBusinessConnection
	UpdateBusinessMessage         = telegram.UpdateBusinessMessage
	UpdateEditedBusinessMessage   = telegram.UpdateEditedBusinessMessage
	UpdateDeletedBusinessMessages = telegram.UpdateDeletedBusinessMessages
	UpdateGuestMessage            = telegram.UpdateGuestMessage
	UpdateMessageReaction         = telegram.UpdateMessageReaction
	UpdateMessageReactionCount    = telegram.UpdateMessageReactionCount
	UpdateInlineQuery             = telegram.UpdateInlineQuery
	UpdateChosenInlineResult      = telegram.UpdateChosenInlineResult
	UpdateCallbackQuery           = telegram.UpdateCallbackQuery
	UpdateShippingQuery           = telegram.UpdateShippingQuery
	UpdatePreCheckoutQuery        = telegram.UpdatePreCheckoutQuery
	UpdatePurchasedPaidMedia      = telegram.UpdatePurchasedPaidMedia
	UpdatePoll                    = telegram.UpdatePoll
	UpdatePollAnswer              = telegram.UpdatePollAnswer
	UpdateMyChatMember            = telegram.UpdateMyChatMember
	UpdateChatMember              = telegram.UpdateChatMember
	UpdateChatJoinRequest         = telegram.UpdateChatJoinRequest
	UpdateChatBoost               = telegram.UpdateChatBoost
	UpdateRemovedChatBoost        = telegram.UpdateRemovedChatBoost
	UpdateManagedBot              = telegram.UpdateManagedBot
	UpdateSubscription            = telegram.UpdateSubscription
)

// EmojiReaction constructs a standard emoji reaction.
func EmojiReaction(emoji string) ReactionType { return telegram.EmojiReaction(emoji) }

// CustomEmojiReaction constructs a custom-emoji reaction by identifier.
func CustomEmojiReaction(id string) ReactionType { return telegram.CustomEmojiReaction(id) }

// PaidReaction constructs Telegram's paid star reaction.
func PaidReaction() ReactionType { return telegram.PaidReaction() }

// AccessibleMessage wraps an ordinary message in Telegram's
// MaybeInaccessibleMessage union.
func AccessibleMessage(message *Message) *MaybeInaccessibleMessage {
	return telegram.AccessibleMessage(message)
}

// UnavailableMessage wraps an inaccessible message in Telegram's
// MaybeInaccessibleMessage union.
func UnavailableMessage(message *InaccessibleMessage) *MaybeInaccessibleMessage {
	return telegram.UnavailableMessage(message)
}

// Button creates an inline callback button.
func Button(text, data string) InlineKeyboardButton { return telegram.Button(text, data) }

// URLButton creates an inline button that opens a URL.
func URLButton(text, url string) InlineKeyboardButton { return telegram.URLButton(text, url) }

// WebAppButton creates an inline button that opens a Telegram Web App.
func WebAppButton(text, url string) InlineKeyboardButton { return telegram.WebAppButton(text, url) }

// CopyButton creates an inline button that copies text.
func CopyButton(text, value string) InlineKeyboardButton { return telegram.CopyButton(text, value) }

// PayButton creates the first-row payment button used with invoices.
func PayButton(text string) InlineKeyboardButton { return telegram.PayButton(text) }

// Keyboard constructs an inline keyboard from rows.
func Keyboard(rows ...[]InlineKeyboardButton) InlineKeyboardMarkup { return telegram.Keyboard(rows...) }

// Row groups inline buttons into one keyboard row.
func Row(buttons ...InlineKeyboardButton) []InlineKeyboardButton { return telegram.Row(buttons...) }

// Key creates a plain reply-keyboard button.
func Key(text string) KeyboardButton { return telegram.Key(text) }

// ReplyKeyboard constructs a reply keyboard from rows.
func ReplyKeyboard(rows ...[]KeyboardButton) ReplyKeyboardMarkup {
	return telegram.ReplyKeyboard(rows...)
}

// KeyRow groups reply-keyboard buttons into one row.
func KeyRow(buttons ...KeyboardButton) []KeyboardButton { return telegram.KeyRow(buttons...) }

// RemoveKeyboard requests removal of the current custom reply keyboard.
func RemoveKeyboard(selective ...bool) ReplyKeyboardRemove {
	return telegram.RemoveKeyboard(selective...)
}

// NewForceReply requests a reply and optionally displays a placeholder.
func NewForceReply(placeholder string) ForceReply { return telegram.NewForceReply(placeholder) }

// DecodeUpdate decodes one update and optionally preserves the original JSON.
func DecodeUpdate(data []byte, preserveRaw bool) (Update, error) {
	return telegram.DecodeUpdate(data, preserveRaw)
}

// DecodeUpdates decodes a getUpdates result and optionally preserves each
// update's original JSON.
func DecodeUpdates(data []byte, preserveRaw bool) ([]Update, error) {
	return telegram.DecodeUpdates(data, preserveRaw)
}
