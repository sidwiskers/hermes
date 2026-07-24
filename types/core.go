package types

import "encoding/json"

// ResponseParameters contains information Telegram attaches to certain errors.
type ResponseParameters struct {
	MigrateToChatID int64 `json:"migrate_to_chat_id,omitempty"`
	RetryAfter      int   `json:"retry_after,omitempty"`
}

// UpdateType identifies the single payload carried by an Update.
type UpdateType string

const (
	UpdateUnknown                 UpdateType = "unknown"
	UpdateMessage                 UpdateType = "message"
	UpdateEditedMessage           UpdateType = "edited_message"
	UpdateChannelPost             UpdateType = "channel_post"
	UpdateEditedChannelPost       UpdateType = "edited_channel_post"
	UpdateBusinessConnection      UpdateType = "business_connection"
	UpdateBusinessMessage         UpdateType = "business_message"
	UpdateEditedBusinessMessage   UpdateType = "edited_business_message"
	UpdateDeletedBusinessMessages UpdateType = "deleted_business_messages"
	UpdateGuestMessage            UpdateType = "guest_message"
	UpdateMessageReaction         UpdateType = "message_reaction"
	UpdateMessageReactionCount    UpdateType = "message_reaction_count"
	UpdateInlineQuery             UpdateType = "inline_query"
	UpdateChosenInlineResult      UpdateType = "chosen_inline_result"
	UpdateCallbackQuery           UpdateType = "callback_query"
	UpdateShippingQuery           UpdateType = "shipping_query"
	UpdatePreCheckoutQuery        UpdateType = "pre_checkout_query"
	UpdatePurchasedPaidMedia      UpdateType = "purchased_paid_media"
	UpdatePoll                    UpdateType = "poll"
	UpdatePollAnswer              UpdateType = "poll_answer"
	UpdateMyChatMember            UpdateType = "my_chat_member"
	UpdateChatMember              UpdateType = "chat_member"
	UpdateChatJoinRequest         UpdateType = "chat_join_request"
	UpdateChatBoost               UpdateType = "chat_boost"
	UpdateRemovedChatBoost        UpdateType = "removed_chat_boost"
	UpdateManagedBot              UpdateType = "managed_bot"
	UpdateSubscription            UpdateType = "subscription"
)

// Update is one incoming Telegram update. Raw is populated only when the
// caller explicitly requests raw preservation through DecodeUpdate or a
// configured API/runtime decoder. The zero-value fast path avoids copying the
// original payload for every update.
type Update struct {
	UpdateID                int64                        `json:"update_id"`
	Message                 *Message                     `json:"message,omitempty"`
	EditedMessage           *Message                     `json:"edited_message,omitempty"`
	ChannelPost             *Message                     `json:"channel_post,omitempty"`
	EditedChannelPost       *Message                     `json:"edited_channel_post,omitempty"`
	BusinessConnection      *BusinessConnection          `json:"business_connection,omitempty"`
	BusinessMessage         *Message                     `json:"business_message,omitempty"`
	EditedBusinessMessage   *Message                     `json:"edited_business_message,omitempty"`
	DeletedBusinessMessages *BusinessMessagesDeleted     `json:"deleted_business_messages,omitempty"`
	GuestMessage            *Message                     `json:"guest_message,omitempty"`
	MessageReaction         *MessageReactionUpdated      `json:"message_reaction,omitempty"`
	MessageReactionCount    *MessageReactionCountUpdated `json:"message_reaction_count,omitempty"`
	InlineQuery             *InlineQuery                 `json:"inline_query,omitempty"`
	ChosenInlineResult      *ChosenInlineResult          `json:"chosen_inline_result,omitempty"`
	CallbackQuery           *CallbackQuery               `json:"callback_query,omitempty"`
	ShippingQuery           *ShippingQuery               `json:"shipping_query,omitempty"`
	PreCheckoutQuery        *PreCheckoutQuery            `json:"pre_checkout_query,omitempty"`
	PurchasedPaidMedia      *PaidMediaPurchased          `json:"purchased_paid_media,omitempty"`
	Poll                    *Poll                        `json:"poll,omitempty"`
	PollAnswer              *PollAnswer                  `json:"poll_answer,omitempty"`
	MyChatMember            *ChatMemberUpdated           `json:"my_chat_member,omitempty"`
	ChatMember              *ChatMemberUpdated           `json:"chat_member,omitempty"`
	ChatJoinRequest         *ChatJoinRequest             `json:"chat_join_request,omitempty"`
	ChatBoost               *ChatBoostUpdated            `json:"chat_boost,omitempty"`
	RemovedChatBoost        *ChatBoostRemoved            `json:"removed_chat_boost,omitempty"`
	ManagedBot              *ManagedBotUpdated           `json:"managed_bot,omitempty"`
	Subscription            *BotSubscriptionUpdated      `json:"subscription,omitempty"`
	Raw                     json.RawMessage              `json:"-"`
}

func (u *Update) Type() UpdateType {
	if u == nil {
		return UpdateUnknown
	}
	switch {
	case u.Message != nil:
		return UpdateMessage
	case u.EditedMessage != nil:
		return UpdateEditedMessage
	case u.ChannelPost != nil:
		return UpdateChannelPost
	case u.EditedChannelPost != nil:
		return UpdateEditedChannelPost
	case u.BusinessConnection != nil:
		return UpdateBusinessConnection
	case u.BusinessMessage != nil:
		return UpdateBusinessMessage
	case u.EditedBusinessMessage != nil:
		return UpdateEditedBusinessMessage
	case u.DeletedBusinessMessages != nil:
		return UpdateDeletedBusinessMessages
	case u.GuestMessage != nil:
		return UpdateGuestMessage
	case u.MessageReaction != nil:
		return UpdateMessageReaction
	case u.MessageReactionCount != nil:
		return UpdateMessageReactionCount
	case u.InlineQuery != nil:
		return UpdateInlineQuery
	case u.ChosenInlineResult != nil:
		return UpdateChosenInlineResult
	case u.CallbackQuery != nil:
		return UpdateCallbackQuery
	case u.ShippingQuery != nil:
		return UpdateShippingQuery
	case u.PreCheckoutQuery != nil:
		return UpdatePreCheckoutQuery
	case u.PurchasedPaidMedia != nil:
		return UpdatePurchasedPaidMedia
	case u.Poll != nil:
		return UpdatePoll
	case u.PollAnswer != nil:
		return UpdatePollAnswer
	case u.MyChatMember != nil:
		return UpdateMyChatMember
	case u.ChatMember != nil:
		return UpdateChatMember
	case u.ChatJoinRequest != nil:
		return UpdateChatJoinRequest
	case u.ChatBoost != nil:
		return UpdateChatBoost
	case u.RemovedChatBoost != nil:
		return UpdateRemovedChatBoost
	case u.ManagedBot != nil:
		return UpdateManagedBot
	case u.Subscription != nil:
		return UpdateSubscription
	default:
		return UpdateUnknown
	}
}

func (u *Update) PrimaryMessage() *Message {
	if u == nil {
		return nil
	}
	switch {
	case u.Message != nil:
		return u.Message
	case u.EditedMessage != nil:
		return u.EditedMessage
	case u.ChannelPost != nil:
		return u.ChannelPost
	case u.EditedChannelPost != nil:
		return u.EditedChannelPost
	case u.BusinessMessage != nil:
		return u.BusinessMessage
	case u.EditedBusinessMessage != nil:
		return u.EditedBusinessMessage
	case u.GuestMessage != nil:
		return u.GuestMessage
	case u.CallbackQuery != nil && u.CallbackQuery.Message != nil:
		return u.CallbackQuery.Message.Message
	default:
		return nil
	}
}

func (u *Update) Sender() *User {
	if u == nil {
		return nil
	}
	switch {
	case u.CallbackQuery != nil:
		return &u.CallbackQuery.From
	case u.BusinessConnection != nil:
		return &u.BusinessConnection.User
	case u.MessageReaction != nil && u.MessageReaction.User != nil:
		return u.MessageReaction.User
	case u.MyChatMember != nil:
		return &u.MyChatMember.From
	case u.ChatMember != nil:
		return &u.ChatMember.From
	case u.InlineQuery != nil:
		return &u.InlineQuery.From
	case u.ChosenInlineResult != nil:
		return &u.ChosenInlineResult.From
	case u.ShippingQuery != nil:
		return &u.ShippingQuery.From
	case u.PreCheckoutQuery != nil:
		return &u.PreCheckoutQuery.From
	case u.PurchasedPaidMedia != nil:
		return &u.PurchasedPaidMedia.From
	case u.PollAnswer != nil && u.PollAnswer.User != nil:
		return u.PollAnswer.User
	case u.ChatJoinRequest != nil:
		return &u.ChatJoinRequest.From
	case u.Subscription != nil:
		return &u.Subscription.User
	case u.ManagedBot != nil:
		return &u.ManagedBot.User
	}
	if message := u.PrimaryMessage(); message != nil {
		return message.From
	}
	return nil
}

type User struct {
	ID                         int64  `json:"id"`
	IsBot                      bool   `json:"is_bot"`
	FirstName                  string `json:"first_name"`
	LastName                   string `json:"last_name,omitempty"`
	Username                   string `json:"username,omitempty"`
	LanguageCode               string `json:"language_code,omitempty"`
	IsPremium                  bool   `json:"is_premium,omitempty"`
	AddedToAttachmentMenu      bool   `json:"added_to_attachment_menu,omitempty"`
	CanJoinGroups              bool   `json:"can_join_groups,omitempty"`
	CanReadAllGroupMessages    bool   `json:"can_read_all_group_messages,omitempty"`
	SupportsGuestQueries       bool   `json:"supports_guest_queries,omitempty"`
	SupportsInlineQueries      bool   `json:"supports_inline_queries,omitempty"`
	CanConnectToBusiness       bool   `json:"can_connect_to_business,omitempty"`
	HasMainWebApp              bool   `json:"has_main_web_app,omitempty"`
	HasTopicsEnabled           bool   `json:"has_topics_enabled,omitempty"`
	AllowsUsersToCreateTopics  bool   `json:"allows_users_to_create_topics,omitempty"`
	CanManageBots              bool   `json:"can_manage_bots,omitempty"`
	SupportsJoinRequestQueries bool   `json:"supports_join_request_queries,omitempty"`
}

type Chat struct {
	ID               int64  `json:"id"`
	Type             string `json:"type"`
	Title            string `json:"title,omitempty"`
	Username         string `json:"username,omitempty"`
	FirstName        string `json:"first_name,omitempty"`
	LastName         string `json:"last_name,omitempty"`
	IsForum          bool   `json:"is_forum,omitempty"`
	IsDirectMessages bool   `json:"is_direct_messages,omitempty"`
}

func (c Chat) IsPrivate() bool    { return c.Type == "private" }
func (c Chat) IsGroup() bool      { return c.Type == "group" || c.Type == "supergroup" }
func (c Chat) IsSupergroup() bool { return c.Type == "supergroup" }
func (c Chat) IsChannel() bool    { return c.Type == "channel" }

// ChatFullInfo is the complete chat metadata returned by getChat.
type ChatFullInfo struct {
	Chat
	ChatFullInfoBotAPIFields
	AccentColorID                int        `json:"accent_color_id"`
	MaxReactionCount             int        `json:"max_reaction_count"`
	ActiveUsernames              []string   `json:"active_usernames,omitempty"`
	Bio                          string     `json:"bio,omitempty"`
	Description                  string     `json:"description,omitempty"`
	InviteLink                   string     `json:"invite_link,omitempty"`
	PinnedMessage                *Message   `json:"pinned_message,omitempty"`
	SlowModeDelay                int        `json:"slow_mode_delay,omitempty"`
	MessageAutoDeleteTime        int        `json:"message_auto_delete_time,omitempty"`
	HasAggressiveAntiSpamEnabled bool       `json:"has_aggressive_anti_spam_enabled,omitempty"`
	HasHiddenMembers             bool       `json:"has_hidden_members,omitempty"`
	HasProtectedContent          bool       `json:"has_protected_content,omitempty"`
	HasVisibleHistory            bool       `json:"has_visible_history,omitempty"`
	StickerSetName               string     `json:"sticker_set_name,omitempty"`
	CanSetStickerSet             bool       `json:"can_set_sticker_set,omitempty"`
	LinkedChatID                 int64      `json:"linked_chat_id,omitempty"`
	Community                    *Community `json:"community,omitempty"`
}

// MessageID is zero for ephemeral messages. EphemeralMessageID is scoped to
// the chat and receiver and may be reused after deletion or expiry.
type Message struct {
	MessageBotAPIFields
	MessageID             int                       `json:"message_id"`
	MessageThreadID       int                       `json:"message_thread_id,omitempty"`
	From                  *User                     `json:"from,omitempty"`
	SenderChat            *Chat                     `json:"sender_chat,omitempty"`
	Date                  int64                     `json:"date"`
	Chat                  Chat                      `json:"chat"`
	ForwardOrigin         *MessageOrigin            `json:"forward_origin,omitempty"`
	ReplyToMessage        *Message                  `json:"reply_to_message,omitempty"`
	ReceiverUser          *User                     `json:"receiver_user,omitempty"`
	EphemeralMessageID    int                       `json:"ephemeral_message_id,omitempty"`
	Text                  string                    `json:"text,omitempty"`
	Caption               string                    `json:"caption,omitempty"`
	Entities              []MessageEntity           `json:"entities,omitempty"`
	CaptionEntities       []MessageEntity           `json:"caption_entities,omitempty"`
	Photo                 []PhotoSize               `json:"photo,omitempty"`
	Animation             *Animation                `json:"animation,omitempty"`
	Audio                 *Audio                    `json:"audio,omitempty"`
	Document              *Document                 `json:"document,omitempty"`
	LivePhoto             *LivePhoto                `json:"live_photo,omitempty"`
	Sticker               *Sticker                  `json:"sticker,omitempty"`
	Video                 *Video                    `json:"video,omitempty"`
	VideoNote             *VideoNote                `json:"video_note,omitempty"`
	Voice                 *Voice                    `json:"voice,omitempty"`
	Contact               *Contact                  `json:"contact,omitempty"`
	Location              *Location                 `json:"location,omitempty"`
	Venue                 *Venue                    `json:"venue,omitempty"`
	Poll                  *Poll                     `json:"poll,omitempty"`
	Dice                  *Dice                     `json:"dice,omitempty"`
	MediaGroupID          string                    `json:"media_group_id,omitempty"`
	NewChatMembers        []User                    `json:"new_chat_members,omitempty"`
	LeftChatMember        *User                     `json:"left_chat_member,omitempty"`
	NewChatTitle          string                    `json:"new_chat_title,omitempty"`
	DeleteChatPhoto       bool                      `json:"delete_chat_photo,omitempty"`
	GroupChatCreated      bool                      `json:"group_chat_created,omitempty"`
	SupergroupChatCreated bool                      `json:"supergroup_chat_created,omitempty"`
	ChannelChatCreated    bool                      `json:"channel_chat_created,omitempty"`
	PinnedMessage         *MaybeInaccessibleMessage `json:"pinned_message,omitempty"`
	ReplyMarkup           *InlineKeyboardMarkup     `json:"reply_markup,omitempty"`
	CommunityChatAdded    *CommunityChatAdded       `json:"community_chat_added,omitempty"`
	CommunityChatRemoved  *CommunityChatRemoved     `json:"community_chat_removed,omitempty"`
	RichMessage           *RichMessage              `json:"rich_message,omitempty"`
	Invoice               *Invoice                  `json:"invoice,omitempty"`
	SuccessfulPayment     *SuccessfulPayment        `json:"successful_payment,omitempty"`
	RefundedPayment       *RefundedPayment          `json:"refunded_payment,omitempty"`
	PaidMedia             *PaidMediaInfo            `json:"paid_media,omitempty"`
	Game                  *Game                     `json:"game,omitempty"`
	ManagedBotCreated     *ManagedBotCreated        `json:"managed_bot_created,omitempty"`
	Checklist             *Checklist                `json:"checklist,omitempty"`
}

func (m *Message) ContentText() string {
	if m == nil {
		return ""
	}
	if m.Text != "" {
		return m.Text
	}
	return m.Caption
}

func (m *Message) IsEphemeral() bool { return m != nil && m.EphemeralMessageID != 0 }

type CallbackQuery struct {
	ID              string                    `json:"id"`
	From            User                      `json:"from"`
	Message         *MaybeInaccessibleMessage `json:"message,omitempty"`
	InlineMessageID string                    `json:"inline_message_id,omitempty"`
	ChatInstance    string                    `json:"chat_instance"`
	Data            string                    `json:"data,omitempty"`
	GameShortName   string                    `json:"game_short_name,omitempty"`
}

type MessageEntity struct {
	MessageEntityBotAPIFields
	Type          string `json:"type"`
	Offset        int    `json:"offset"`
	Length        int    `json:"length"`
	URL           string `json:"url,omitempty"`
	User          *User  `json:"user,omitempty"`
	Language      string `json:"language,omitempty"`
	CustomEmojiID string `json:"custom_emoji_id,omitempty"`
}

type ReplyParameters struct {
	ReplyParametersBotAPIFields
	MessageID                int             `json:"message_id,omitempty"`
	ChatID                   any             `json:"chat_id,omitempty"`
	AllowSendingWithoutReply bool            `json:"allow_sending_without_reply,omitempty"`
	Quote                    string          `json:"quote,omitempty"`
	QuoteParseMode           string          `json:"quote_parse_mode,omitempty"`
	QuoteEntities            []MessageEntity `json:"quote_entities,omitempty"`
	QuotePosition            int             `json:"quote_position,omitempty"`
	EphemeralMessageID       int             `json:"ephemeral_message_id,omitempty"`
}

type LinkPreviewOptions struct {
	IsDisabled       bool   `json:"is_disabled,omitempty"`
	URL              string `json:"url,omitempty"`
	PreferSmallMedia bool   `json:"prefer_small_media,omitempty"`
	PreferLargeMedia bool   `json:"prefer_large_media,omitempty"`
	ShowAboveText    bool   `json:"show_above_text,omitempty"`
}

// EphemeralMessageRef uniquely addresses an ephemeral message for its current
// lifetime. Telegram may reuse EphemeralMessageID after deletion or expiry.
type EphemeralMessageRef struct {
	ChatID             any
	ReceiverUserID     int64
	EphemeralMessageID int
}

// EphemeralRef returns the identifiers required by ephemeral edit/delete methods.
func (m *Message) EphemeralRef() (EphemeralMessageRef, bool) {
	if m == nil || m.EphemeralMessageID == 0 || m.ReceiverUser == nil {
		return EphemeralMessageRef{}, false
	}
	return EphemeralMessageRef{
		ChatID:             m.Chat.ID,
		ReceiverUserID:     m.ReceiverUser.ID,
		EphemeralMessageID: m.EphemeralMessageID,
	}, true
}
