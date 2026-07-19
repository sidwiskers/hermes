package framework

import "strings"

// Filter decides whether an update route applies.
type Filter func(*Context) bool

// All matches when every non-nil filter matches.
func All(filters ...Filter) Filter {
	return func(c *Context) bool {
		for _, filter := range filters {
			if filter != nil && !filter(c) {
				return false
			}
		}
		return true
	}
}

// Any matches when at least one non-nil filter matches.
func Any(filters ...Filter) Filter {
	return func(c *Context) bool {
		for _, filter := range filters {
			if filter != nil && filter(c) {
				return true
			}
		}
		return false
	}
}

// Not negates filter. A nil filter is treated as false before negation.
func Not(filter Filter) Filter {
	return func(c *Context) bool { return filter == nil || !filter(c) }
}

// UpdateIs matches any supplied update type.
func UpdateIs(types ...UpdateType) Filter {
	set := make(map[UpdateType]struct{}, len(types))
	for _, typ := range types {
		set[typ] = struct{}{}
	}
	return func(c *Context) bool {
		if c == nil || c.Update == nil {
			return false
		}
		_, ok := set[c.Update.Type()]
		return ok
	}
}

// MessageUpdate matches updates with a primary message.
func MessageUpdate(c *Context) bool { return c != nil && c.Message != nil }

// CallbackUpdate matches callback-query updates.
func CallbackUpdate(c *Context) bool { return c != nil && c.Callback != nil }

// TextMessage matches messages containing text.
func TextMessage(c *Context) bool { return c != nil && c.Message != nil && c.Message.Text != "" }

// CaptionedMessage matches messages containing a caption.
func CaptionedMessage(c *Context) bool {
	return c != nil && c.Message != nil && c.Message.Caption != ""
}

// PhotoMessage matches messages containing photos.
func PhotoMessage(c *Context) bool {
	return c != nil && c.Message != nil && len(c.Message.Photo) != 0
}

// DocumentMessage matches messages containing a document.
func DocumentMessage(c *Context) bool {
	return c != nil && c.Message != nil && c.Message.Document != nil
}

// StickerMessage matches messages containing a sticker.
func StickerMessage(c *Context) bool {
	return c != nil && c.Message != nil && c.Message.Sticker != nil
}

// VideoMessage matches messages containing a video.
func VideoMessage(c *Context) bool {
	return c != nil && c.Message != nil && c.Message.Video != nil
}

// VoiceMessage matches messages containing a voice note.
func VoiceMessage(c *Context) bool {
	return c != nil && c.Message != nil && c.Message.Voice != nil
}

// PrivateChat matches messages from private chats.
func PrivateChat(c *Context) bool {
	return c != nil && c.Message != nil && c.Message.Chat.IsPrivate()
}

// GroupChat matches messages from groups and supergroups.
func GroupChat(c *Context) bool {
	return c != nil && c.Message != nil && c.Message.Chat.IsGroup()
}

// ChannelChat matches channel messages.
func ChannelChat(c *Context) bool {
	return c != nil && c.Message != nil && c.Message.Chat.IsChannel()
}

// EphemeralMessage matches Telegram ephemeral messages.
func EphemeralMessage(c *Context) bool {
	return c != nil && c.Message != nil && c.Message.IsEphemeral()
}

// FromUsers matches updates sent by any supplied user ID.
func FromUsers(ids ...int64) Filter {
	set := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return func(c *Context) bool {
		user := c.Sender()
		if user == nil {
			return false
		}
		_, ok := set[user.ID]
		return ok
	}
}

// InChats matches updates belonging to any supplied chat ID.
func InChats(ids ...int64) Filter {
	set := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return func(c *Context) bool {
		id, ok := c.ChatID()
		if !ok {
			return false
		}
		_, ok = set[id]
		return ok
	}
}

// TextEquals matches primary message text or captions exactly.
func TextEquals(values ...string) Filter {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return func(c *Context) bool {
		if c == nil || c.Message == nil {
			return false
		}
		_, ok := set[c.Message.ContentText()]
		return ok
	}
}

// TextPrefix matches primary message text or captions by prefix.
func TextPrefix(prefixes ...string) Filter {
	return func(c *Context) bool {
		if c == nil || c.Message == nil {
			return false
		}
		text := c.Message.ContentText()
		for _, prefix := range prefixes {
			if strings.HasPrefix(text, prefix) {
				return true
			}
		}
		return false
	}
}

// CallbackDataPrefix matches callback data by prefix.
func CallbackDataPrefix(prefixes ...string) Filter {
	return func(c *Context) bool {
		if c == nil || c.Callback == nil {
			return false
		}
		for _, prefix := range prefixes {
			if strings.HasPrefix(c.Callback.Data, prefix) {
				return true
			}
		}
		return false
	}
}
