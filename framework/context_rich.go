package framework

import "fmt"

// Rich sends a persistent rich message to the current chat.
func (c *Context) Rich(message InputRichMessage, options ...SendOption) error {
	_, err := c.RichMessage(message, options...)
	return err
}

// RichMessage sends a persistent rich message and returns Telegram's message.
func (c *Context) RichMessage(message InputRichMessage, options ...SendOption) (*Message, error) {
	chatID, ok := c.ChatID()
	if !ok {
		return nil, fmt.Errorf("hermes: update has no chat")
	}
	resolved := resolveSendOptions(options)
	return c.Bot.SendRichMessage(c.Context, SendRichMessageParams{
		ChatID:              chatID,
		MessageThreadID:     resolved.messageThreadID,
		RichMessage:         message,
		DisableNotification: resolved.disableNotification,
		ProtectContent:      resolved.protectContent,
		MessageEffectID:     resolved.messageEffectID,
		ReplyMarkup:         resolved.replyMarkup,
	})
}
