package types

// ReplyMarkup is implemented by all Telegram reply-markup objects.
// The marker prevents accidental use of unrelated values in high-level helpers.
type ReplyMarkup interface {
	replyMarkup()
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

func (InlineKeyboardMarkup) replyMarkup() {}

type InlineKeyboardButton struct {
	InlineKeyboardButtonBotAPIFields
	Text                         string                       `json:"text"`
	URL                          string                       `json:"url,omitempty"`
	CallbackData                 string                       `json:"callback_data,omitempty"`
	WebApp                       *WebAppInfo                  `json:"web_app,omitempty"`
	LoginURL                     *LoginURL                    `json:"login_url,omitempty"`
	SwitchInlineQuery            string                       `json:"switch_inline_query,omitempty"`
	SwitchInlineQueryCurrentChat string                       `json:"switch_inline_query_current_chat,omitempty"`
	SwitchInlineQueryChosenChat  *SwitchInlineQueryChosenChat `json:"switch_inline_query_chosen_chat,omitempty"`
	CopyText                     *CopyTextButton              `json:"copy_text,omitempty"`
	CallbackGame                 *CallbackGame                `json:"callback_game,omitempty"`
	Pay                          bool                         `json:"pay,omitempty"`
}

type WebAppInfo struct {
	URL string `json:"url"`
}

type LoginURL struct {
	URL                string `json:"url"`
	ForwardText        string `json:"forward_text,omitempty"`
	BotUsername        string `json:"bot_username,omitempty"`
	RequestWriteAccess bool   `json:"request_write_access,omitempty"`
}

type SwitchInlineQueryChosenChat struct {
	Query             string `json:"query,omitempty"`
	AllowUserChats    bool   `json:"allow_user_chats,omitempty"`
	AllowBotChats     bool   `json:"allow_bot_chats,omitempty"`
	AllowGroupChats   bool   `json:"allow_group_chats,omitempty"`
	AllowChannelChats bool   `json:"allow_channel_chats,omitempty"`
}

type CopyTextButton struct {
	Text string `json:"text"`
}

type CallbackGame struct{}

func Button(text, data string) InlineKeyboardButton {
	return InlineKeyboardButton{Text: text, CallbackData: data}
}

func URLButton(text, url string) InlineKeyboardButton {
	return InlineKeyboardButton{Text: text, URL: url}
}

func WebAppButton(text, url string) InlineKeyboardButton {
	return InlineKeyboardButton{Text: text, WebApp: &WebAppInfo{URL: url}}
}

func CopyButton(text, value string) InlineKeyboardButton {
	return InlineKeyboardButton{Text: text, CopyText: &CopyTextButton{Text: value}}
}

func PayButton(text string) InlineKeyboardButton {
	return InlineKeyboardButton{Text: text, Pay: true}
}

func Keyboard(rows ...[]InlineKeyboardButton) InlineKeyboardMarkup {
	return InlineKeyboardMarkup{InlineKeyboard: rows}
}

func Row(buttons ...InlineKeyboardButton) []InlineKeyboardButton { return buttons }

type ReplyKeyboardMarkup struct {
	Keyboard              [][]KeyboardButton `json:"keyboard"`
	IsPersistent          bool               `json:"is_persistent,omitempty"`
	ResizeKeyboard        bool               `json:"resize_keyboard,omitempty"`
	OneTimeKeyboard       bool               `json:"one_time_keyboard,omitempty"`
	InputFieldPlaceholder string             `json:"input_field_placeholder,omitempty"`
	Selective             bool               `json:"selective,omitempty"`
}

func (ReplyKeyboardMarkup) replyMarkup() {}

type KeyboardButton struct {
	KeyboardButtonBotAPIFields
	Text              string                           `json:"text"`
	RequestUsers      *KeyboardButtonRequestUsers      `json:"request_users,omitempty"`
	RequestChat       *KeyboardButtonRequestChat       `json:"request_chat,omitempty"`
	RequestManagedBot *KeyboardButtonRequestManagedBot `json:"request_managed_bot,omitempty"`
	RequestContact    bool                             `json:"request_contact,omitempty"`
	RequestLocation   bool                             `json:"request_location,omitempty"`
	RequestPoll       *KeyboardButtonPollType          `json:"request_poll,omitempty"`
	WebApp            *WebAppInfo                      `json:"web_app,omitempty"`
}

type KeyboardButtonRequestUsers struct {
	RequestID       int   `json:"request_id"`
	UserIsBot       *bool `json:"user_is_bot,omitempty"`
	UserIsPremium   *bool `json:"user_is_premium,omitempty"`
	MaxQuantity     int   `json:"max_quantity,omitempty"`
	RequestName     bool  `json:"request_name,omitempty"`
	RequestUsername bool  `json:"request_username,omitempty"`
	RequestPhoto    bool  `json:"request_photo,omitempty"`
}

type KeyboardButtonRequestChat struct {
	KeyboardButtonRequestChatBotAPIFields
	RequestID       int   `json:"request_id"`
	ChatIsChannel   bool  `json:"chat_is_channel"`
	ChatIsForum     *bool `json:"chat_is_forum,omitempty"`
	ChatHasUsername *bool `json:"chat_has_username,omitempty"`
	ChatIsCreated   bool  `json:"chat_is_created,omitempty"`
	BotIsMember     bool  `json:"bot_is_member,omitempty"`
	RequestTitle    bool  `json:"request_title,omitempty"`
	RequestUsername bool  `json:"request_username,omitempty"`
	RequestPhoto    bool  `json:"request_photo,omitempty"`
}

type KeyboardButtonRequestManagedBot struct {
	RequestID         int    `json:"request_id"`
	SuggestedName     string `json:"suggested_name,omitempty"`
	SuggestedUsername string `json:"suggested_username,omitempty"`
}

type KeyboardButtonPollType struct {
	Type string `json:"type,omitempty"`
}

func Key(text string) KeyboardButton { return KeyboardButton{Text: text} }

func ReplyKeyboard(rows ...[]KeyboardButton) ReplyKeyboardMarkup {
	return ReplyKeyboardMarkup{Keyboard: rows, ResizeKeyboard: true}
}

func KeyRow(buttons ...KeyboardButton) []KeyboardButton { return buttons }

type ReplyKeyboardRemove struct {
	RemoveKeyboard bool `json:"remove_keyboard"`
	Selective      bool `json:"selective,omitempty"`
}

func (ReplyKeyboardRemove) replyMarkup() {}

func RemoveKeyboard(selective ...bool) ReplyKeyboardRemove {
	value := false
	if len(selective) != 0 {
		value = selective[0]
	}
	return ReplyKeyboardRemove{RemoveKeyboard: true, Selective: value}
}

type ForceReply struct {
	ForceReply            bool   `json:"force_reply"`
	InputFieldPlaceholder string `json:"input_field_placeholder,omitempty"`
	Selective             bool   `json:"selective,omitempty"`
}

func (ForceReply) replyMarkup() {}

// ForceReplyMarkup is retained as a compatibility alias.
type ForceReplyMarkup = ForceReply

func NewForceReply(placeholder string) ForceReply {
	return ForceReply{ForceReply: true, InputFieldPlaceholder: placeholder}
}
