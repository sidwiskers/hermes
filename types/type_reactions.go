package types

const (
	ReactionEmoji       = "emoji"
	ReactionCustomEmoji = "custom_emoji"
	ReactionPaid        = "paid"
)

// ReactionType represents an emoji, custom emoji, or paid reaction.
type ReactionType struct {
	Type          string `json:"type"`
	Emoji         string `json:"emoji,omitempty"`
	CustomEmojiID string `json:"custom_emoji_id,omitempty"`
}

func EmojiReaction(emoji string) ReactionType {
	return ReactionType{Type: ReactionEmoji, Emoji: emoji}
}

func CustomEmojiReaction(id string) ReactionType {
	return ReactionType{Type: ReactionCustomEmoji, CustomEmojiID: id}
}

func PaidReaction() ReactionType { return ReactionType{Type: ReactionPaid} }

type ReactionCount struct {
	Type       ReactionType `json:"type"`
	TotalCount int          `json:"total_count"`
}
