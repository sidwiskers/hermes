package framework

import "fmt"

// Edit replaces text on the current ordinary, inline, or ephemeral message.
func (c *Context) Edit(text string, options ...SendOption) error {
	if c == nil || c.Bot == nil {
		return fmt.Errorf("hermes: nil context")
	}
	resolved := resolveSendOptions(options)
	if c.Message != nil && c.Message.IsEphemeral() {
		ref, err := c.ephemeralRef()
		if err != nil {
			return err
		}
		return c.Bot.EditEphemeralText(c.Context, EditEphemeralMessageTextParams{
			ChatID: ref.ChatID, ReceiverUserID: ref.ReceiverUserID, EphemeralMessageID: ref.EphemeralMessageID,
			Text: text, ParseMode: resolved.parseMode, LinkPreviewOptions: previewOptions(resolved), ReplyMarkup: inlineMarkup(resolved.replyMarkup),
		})
	}
	params := EditMessageTextParams{Text: text, ParseMode: resolved.parseMode, LinkPreviewOptions: previewOptions(resolved), ReplyMarkup: inlineMarkup(resolved.replyMarkup)}
	switch {
	case c.Callback != nil && c.Callback.InlineMessageID != "":
		params.InlineMessageID = c.Callback.InlineMessageID
	case c.Message != nil:
		params.ChatID, params.MessageID = c.Message.Chat.ID, c.Message.MessageID
	default:
		return fmt.Errorf("hermes: update has no editable message")
	}
	_, err := c.Bot.EditMessageText(c.Context, params)
	return err
}

// EditCaption replaces the caption on the current message.
func (c *Context) EditCaption(caption string, options ...SendOption) error {
	resolved := resolveSendOptions(options)
	if c != nil && c.Message != nil && c.Message.IsEphemeral() {
		ref, err := c.ephemeralRef()
		if err != nil {
			return err
		}
		return c.Bot.EditEphemeralCaption(c.Context, EditEphemeralMessageCaptionParams{ChatID: ref.ChatID, ReceiverUserID: ref.ReceiverUserID, EphemeralMessageID: ref.EphemeralMessageID, Caption: caption, ParseMode: resolved.parseMode, ReplyMarkup: inlineMarkup(resolved.replyMarkup)})
	}
	params := EditMessageCaptionParams{Caption: caption, ParseMode: resolved.parseMode, ReplyMarkup: inlineMarkup(resolved.replyMarkup), ShowCaptionAboveMedia: resolved.showCaptionAbove}
	switch {
	case c != nil && c.Callback != nil && c.Callback.InlineMessageID != "":
		params.InlineMessageID = c.Callback.InlineMessageID
	case c != nil && c.Message != nil:
		params.ChatID, params.MessageID = c.Message.Chat.ID, c.Message.MessageID
	default:
		return fmt.Errorf("hermes: update has no editable message")
	}
	_, err := c.Bot.EditMessageCaption(c.Context, params)
	return err
}

// EditKeyboard replaces or removes the current message's inline keyboard.
func (c *Context) EditKeyboard(markup *InlineKeyboardMarkup) error {
	if c != nil && c.Message != nil && c.Message.IsEphemeral() {
		ref, err := c.ephemeralRef()
		if err != nil {
			return err
		}
		return c.Bot.EditEphemeralReplyMarkup(c.Context, EditEphemeralMessageReplyMarkupParams{ChatID: ref.ChatID, ReceiverUserID: ref.ReceiverUserID, EphemeralMessageID: ref.EphemeralMessageID, ReplyMarkup: markup})
	}
	params := EditMessageReplyMarkupParams{ReplyMarkup: markup}
	switch {
	case c != nil && c.Callback != nil && c.Callback.InlineMessageID != "":
		params.InlineMessageID = c.Callback.InlineMessageID
	case c != nil && c.Message != nil:
		params.ChatID, params.MessageID = c.Message.Chat.ID, c.Message.MessageID
	default:
		return fmt.Errorf("hermes: update has no editable message")
	}
	_, err := c.Bot.EditMessageReplyMarkup(c.Context, params)
	return err
}

// Delete removes the current ordinary or ephemeral message.
func (c *Context) Delete() error {
	if c == nil || c.Message == nil {
		return fmt.Errorf("hermes: update has no deletable message")
	}
	if c.Message.IsEphemeral() {
		ref, err := c.ephemeralRef()
		if err != nil {
			return err
		}
		return c.Bot.DeleteEphemeral(c.Context, DeleteEphemeralMessageParams{ChatID: ref.ChatID, ReceiverUserID: ref.ReceiverUserID, EphemeralMessageID: ref.EphemeralMessageID})
	}
	return c.Bot.DeleteMessage(c.Context, DeleteMessageParams{ChatID: c.Message.Chat.ID, MessageID: c.Message.MessageID})
}

func (c *Context) ephemeralRef() (EphemeralMessageRef, error) {
	if c == nil || c.Message == nil || c.Message.EphemeralMessageID == 0 {
		return EphemeralMessageRef{}, fmt.Errorf("hermes: update has no ephemeral message")
	}
	receiver := c.Message.ReceiverUser
	if receiver == nil {
		receiver = c.Sender()
	}
	if receiver == nil {
		return EphemeralMessageRef{}, fmt.Errorf("hermes: ephemeral receiver is unavailable")
	}
	return EphemeralMessageRef{ChatID: c.Message.Chat.ID, ReceiverUserID: receiver.ID, EphemeralMessageID: c.Message.EphemeralMessageID}, nil
}
func previewOptions(options sendOptions) *LinkPreviewOptions {
	if options.noPreview {
		return &LinkPreviewOptions{IsDisabled: true}
	}
	return nil
}
func inlineMarkup(markup ReplyMarkup) *InlineKeyboardMarkup {
	switch value := markup.(type) {
	case InlineKeyboardMarkup:
		return &value
	case *InlineKeyboardMarkup:
		return value
	default:
		return nil
	}
}
