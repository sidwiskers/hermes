package api

import telegram "github.com/sidwiskers/hermes/types"

type (
	ResponseParameters              = telegram.ResponseParameters
	Update                          = telegram.Update
	User                            = telegram.User
	Chat                            = telegram.Chat
	ChatFullInfo                    = telegram.ChatFullInfo
	Message                         = telegram.Message
	CallbackQuery                   = telegram.CallbackQuery
	MessageEntity                   = telegram.MessageEntity
	ReplyParameters                 = telegram.ReplyParameters
	LinkPreviewOptions              = telegram.LinkPreviewOptions
	BotCommand                      = telegram.BotCommand
	File                            = telegram.File
	Sticker                         = telegram.Sticker
	StickerSet                      = telegram.StickerSet
	MaskPosition                    = telegram.MaskPosition
	Location                        = telegram.Location
	ReplyMarkup                     = telegram.ReplyMarkup
	InlineKeyboardMarkup            = telegram.InlineKeyboardMarkup
	KeyboardButton                  = telegram.KeyboardButton
	KeyboardButtonRequestUsers      = telegram.KeyboardButtonRequestUsers
	KeyboardButtonRequestChat       = telegram.KeyboardButtonRequestChat
	KeyboardButtonRequestManagedBot = telegram.KeyboardButtonRequestManagedBot
	EphemeralMessageRef             = telegram.EphemeralMessageRef
	Poll                            = telegram.Poll
	ChatPermissions                 = telegram.ChatPermissions
	ChatMember                      = telegram.ChatMember
	ChatInviteLink                  = telegram.ChatInviteLink
	ReactionType                    = telegram.ReactionType
	RichText                        = telegram.RichText
	RichMessage                     = telegram.RichMessage
	RichBlock                       = telegram.RichBlock
	RichBlockCaption                = telegram.RichBlockCaption
	RichBlockTableCell              = telegram.RichBlockTableCell
	LabeledPrice                    = telegram.LabeledPrice
	ShippingOption                  = telegram.ShippingOption
	StarAmount                      = telegram.StarAmount
	StarTransactions                = telegram.StarTransactions
	ForumTopic                      = telegram.ForumTopic
	BotName                         = telegram.BotName
	BotDescription                  = telegram.BotDescription
	BotShortDescription             = telegram.BotShortDescription
	UserProfilePhotos               = telegram.UserProfilePhotos
	UserProfileAudios               = telegram.UserProfileAudios
	ChatAdministratorRights         = telegram.ChatAdministratorRights
	MenuButton                      = telegram.MenuButton
	BusinessBotRights               = telegram.BusinessBotRights
	BusinessConnection              = telegram.BusinessConnection
	ChatBoostSource                 = telegram.ChatBoostSource
	ChatBoost                       = telegram.ChatBoost
	UserChatBoosts                  = telegram.UserChatBoosts
	ChatBoostUpdated                = telegram.ChatBoostUpdated
	ChatBoostRemoved                = telegram.ChatBoostRemoved
	InlineQueryResultsButton        = telegram.InlineQueryResultsButton
	InlineQueryResultRaw            = telegram.InlineQueryResultRaw
	SentWebAppMessage               = telegram.SentWebAppMessage
	PreparedInlineMessage           = telegram.PreparedInlineMessage
	PreparedKeyboardButton          = telegram.PreparedKeyboardButton
	SentGuestMessage                = telegram.SentGuestMessage
	PassportElementErrorRaw         = telegram.PassportElementErrorRaw
	Game                            = telegram.Game
	GameHighScore                   = telegram.GameHighScore
	BotAccessSettings               = telegram.BotAccessSettings
	ManagedBotCreated               = telegram.ManagedBotCreated
	ManagedBotUpdated               = telegram.ManagedBotUpdated
	AcceptedGiftTypes               = telegram.AcceptedGiftTypes
	Gifts                           = telegram.Gifts
	UniqueGift                      = telegram.UniqueGift
	OwnedGift                       = telegram.OwnedGift
	OwnedGifts                      = telegram.OwnedGifts
	Story                           = telegram.Story
	ChecklistTask                   = telegram.ChecklistTask
	Checklist                       = telegram.Checklist
)

const (
	ReactionEmoji          = telegram.ReactionEmoji
	ReactionCustomEmoji    = telegram.ReactionCustomEmoji
	ReactionPaid           = telegram.ReactionPaid
	MenuButtonTypeCommands = telegram.MenuButtonTypeCommands
	MenuButtonTypeWebApp   = telegram.MenuButtonTypeWebApp
	MenuButtonTypeDefault  = telegram.MenuButtonTypeDefault
)

func CommandsMenuButton() MenuButton { return telegram.CommandsMenuButton() }

func WebAppMenuButton(text, url string) MenuButton {
	return telegram.WebAppMenuButton(text, url)
}

func DefaultMenuButton() MenuButton { return telegram.DefaultMenuButton() }
