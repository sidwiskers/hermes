package types

type BusinessConnection struct {
	ID         string             `json:"id"`
	User       User               `json:"user"`
	UserChatID int64              `json:"user_chat_id"`
	Date       int64              `json:"date"`
	Rights     *BusinessBotRights `json:"rights,omitempty"`
	IsEnabled  bool               `json:"is_enabled"`

	// CanReply is retained for decoding responses from older Bot API servers.
	// Current servers expose the value through Rights.CanReply.
	CanReply bool `json:"can_reply,omitempty"`
}

type BusinessMessagesDeleted struct {
	BusinessConnectionID string `json:"business_connection_id"`
	Chat                 Chat   `json:"chat"`
	MessageIDs           []int  `json:"message_ids"`
}

type ChatMemberUpdated struct {
	ChatMemberUpdatedBotAPIFields
	Chat          Chat       `json:"chat"`
	From          User       `json:"from"`
	Date          int64      `json:"date"`
	OldChatMember ChatMember `json:"old_chat_member"`
	NewChatMember ChatMember `json:"new_chat_member"`
}

type ChatJoinRequest struct {
	ChatJoinRequestBotAPIFields
	Chat       Chat   `json:"chat"`
	From       User   `json:"from"`
	UserChatID int64  `json:"user_chat_id"`
	Date       int64  `json:"date"`
	Bio        string `json:"bio,omitempty"`
	QueryID    string `json:"query_id,omitempty"`
}

type MessageReactionUpdated struct {
	Chat        Chat           `json:"chat"`
	MessageID   int            `json:"message_id"`
	User        *User          `json:"user,omitempty"`
	ActorChat   *Chat          `json:"actor_chat,omitempty"`
	Date        int64          `json:"date"`
	OldReaction []ReactionType `json:"old_reaction"`
	NewReaction []ReactionType `json:"new_reaction"`
}

type MessageReactionCountUpdated struct {
	Chat      Chat            `json:"chat"`
	MessageID int             `json:"message_id"`
	Date      int64           `json:"date"`
	Reactions []ReactionCount `json:"reactions"`
}

type Community struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CommunityChatAdded struct {
	Community Community `json:"community"`
}

type CommunityChatRemoved struct{}

type BotSubscriptionUpdated struct {
	User           User   `json:"user"`
	InvoicePayload string `json:"invoice_payload"`
	State          string `json:"state"`
}
