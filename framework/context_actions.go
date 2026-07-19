package framework

import "fmt"

// React sets one reaction on the current message. Pass an empty ReactionType
// to remove the bot's reaction.
func (c *Context) React(reaction ReactionType, big ...bool) error {
	chatID, ok := c.ChatID()
	if !ok || c.Message == nil || c.Message.MessageID == 0 {
		return fmt.Errorf("hermes: update has no reactable message")
	}
	params := SetMessageReactionParams{ChatID: chatID, MessageID: c.Message.MessageID}
	if reaction.Type != "" {
		params.Reaction = []ReactionType{reaction}
	}
	params.IsBig = len(big) != 0 && big[0]
	return c.Bot.SetMessageReaction(c.Context, params)
}

// Dice sends a Telegram dice and discards the returned message.
func (c *Context) Dice(emoji string, options ...SendOption) error {
	_, err := c.DiceMessage(emoji, options...)
	return err
}

// DiceMessage sends a Telegram dice and returns Telegram's message.
func (c *Context) DiceMessage(emoji string, options ...SendOption) (*Message, error) {
	chatID, ok := c.ChatID()
	if !ok {
		return nil, fmt.Errorf("hermes: update has no chat")
	}
	resolved := resolveSendOptions(options)
	params := SendDiceParams{
		ChatID: chatID, Emoji: emoji,
		MessageThreadID:     resolved.messageThreadID,
		DisableNotification: resolved.disableNotification,
		ProtectContent:      resolved.protectContent,
		MessageEffectID:     resolved.messageEffectID,
		ReplyMarkup:         resolved.replyMarkup,
	}
	return c.Bot.SendDice(c.Context, params)
}

// Poll sends a simple text poll. Use Bot.SendPoll for the complete typed API.
func (c *Context) Poll(question string, answers ...string) error {
	_, err := c.PollMessage(question, answers...)
	return err
}

// PollMessage sends a simple text poll and returns Telegram's message.
func (c *Context) PollMessage(question string, answers ...string) (*Message, error) {
	chatID, ok := c.ChatID()
	if !ok {
		return nil, fmt.Errorf("hermes: update has no chat")
	}
	options := make([]InputPollOption, len(answers))
	for index, answer := range answers {
		options[index] = InputPollOption{Text: answer}
	}
	return c.Bot.SendPoll(c.Context, SendPollParams{ChatID: chatID, Question: question, Options: options})
}

// BanSender bans the current sender from the current chat.
func (c *Context) BanSender(revokeMessages ...bool) error {
	chatID, ok := c.ChatID()
	sender := c.Sender()
	if !ok || sender == nil {
		return fmt.Errorf("hermes: update has no chat sender")
	}
	return c.Bot.BanChatMember(c.Context, BanChatMemberParams{
		ChatID: chatID, UserID: sender.ID,
		RevokeMessages: len(revokeMessages) != 0 && revokeMessages[0],
	})
}

// ApproveJoinRequest approves the current chat-join request.
func (c *Context) ApproveJoinRequest() error {
	if c == nil || c.Update == nil || c.Update.ChatJoinRequest == nil {
		return fmt.Errorf("hermes: update has no chat join request")
	}
	request := c.Update.ChatJoinRequest
	return c.Bot.ApproveChatJoinRequest(c.Context, ChatJoinRequestParams{ChatID: request.Chat.ID, UserID: request.From.ID})
}

// DeclineJoinRequest declines the current chat-join request.
func (c *Context) DeclineJoinRequest() error {
	if c == nil || c.Update == nil || c.Update.ChatJoinRequest == nil {
		return fmt.Errorf("hermes: update has no chat join request")
	}
	request := c.Update.ChatJoinRequest
	return c.Bot.DeclineChatJoinRequest(c.Context, ChatJoinRequestParams{ChatID: request.Chat.ID, UserID: request.From.ID})
}

// Pin pins the current message.
func (c *Context) Pin(options ...SendOption) error {
	chatID, ok := c.ChatID()
	messageID, hasMessage := c.MessageID()
	if !ok || !hasMessage {
		return fmt.Errorf("hermes: update has no pinnable message")
	}
	resolved := resolveSendOptions(options)
	return c.Bot.PinChatMessage(c.Context, PinChatMessageParams{
		ChatID: chatID, MessageID: messageID, DisableNotification: resolved.disableNotification,
	})
}

// Unpin unpins the current message.
func (c *Context) Unpin() error {
	chatID, ok := c.ChatID()
	messageID, hasMessage := c.MessageID()
	if !ok || !hasMessage {
		return fmt.Errorf("hermes: update has no unpinnable message")
	}
	return c.Bot.UnpinChatMessage(c.Context, UnpinChatMessageParams{ChatID: chatID, MessageID: messageID})
}
