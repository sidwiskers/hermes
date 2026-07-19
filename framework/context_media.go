package framework

import "fmt"

type contextSendTarget struct {
	chatID     int64
	receiverID int64
	callbackID string
	reply      *ReplyParameters
	options    sendOptions
}

func (c *Context) mediaTarget(ephemeral bool, options []SendOption) (contextSendTarget, error) {
	chatID, ok := c.ChatID()
	if !ok {
		return contextSendTarget{}, fmt.Errorf("hermes: update has no chat")
	}
	target := contextSendTarget{chatID: chatID, options: resolveSendOptions(options)}
	if ephemeral {
		sender := c.Sender()
		if sender == nil {
			return contextSendTarget{}, fmt.Errorf("hermes: update has no sender")
		}
		target.receiverID = sender.ID
		c.applyEphemeralTarget(&target.callbackID, &target.reply)
	}
	return target, nil
}

func applyPhotoOptions(p *SendPhotoParams, target contextSendTarget) {
	p.ChatID = target.chatID
	p.ParseMode = target.options.parseMode
	p.DisableNotification = target.options.disableNotification
	p.ProtectContent = target.options.protectContent
	p.MessageThreadID = target.options.messageThreadID
	p.ReplyMarkup = target.options.replyMarkup
	p.MessageEffectID = target.options.messageEffectID
	p.HasSpoiler = target.options.hasSpoiler
	p.ShowCaptionAboveMedia = target.options.showCaptionAbove
	p.ReceiverUserID = target.receiverID
	p.CallbackQueryID = target.callbackID
	p.ReplyParameters = target.reply
}

// Photo sends a photo by file ID or URL and discards the returned message.
func (c *Context) Photo(file, caption string, options ...SendOption) error {
	_, err := c.PhotoMessage(file, caption, options...)
	return err
}

// PhotoMessage sends a photo by file ID or URL and returns Telegram's message.
func (c *Context) PhotoMessage(file, caption string, options ...SendOption) (*Message, error) {
	target, err := c.mediaTarget(false, options)
	if err != nil {
		return nil, err
	}
	p := SendPhotoParams{Photo: file, Caption: caption}
	applyPhotoOptions(&p, target)
	return c.Bot.SendPhoto(c.Context, p)
}

// EphemeralPhoto sends a photo visible only to the current sender.
func (c *Context) EphemeralPhoto(file, caption string, options ...SendOption) error {
	_, err := c.EphemeralPhotoMessage(file, caption, options...)
	return err
}

// EphemeralPhotoMessage sends a private photo and returns Telegram's message.
func (c *Context) EphemeralPhotoMessage(file, caption string, options ...SendOption) (*Message, error) {
	target, err := c.mediaTarget(true, options)
	if err != nil {
		return nil, err
	}
	p := SendPhotoParams{Photo: file, Caption: caption}
	applyPhotoOptions(&p, target)
	return c.Bot.SendPhoto(c.Context, p)
}

func applyDocumentOptions(p *SendDocumentParams, target contextSendTarget) {
	p.ChatID = target.chatID
	p.ParseMode = target.options.parseMode
	p.DisableNotification = target.options.disableNotification
	p.ProtectContent = target.options.protectContent
	p.MessageThreadID = target.options.messageThreadID
	p.ReplyMarkup = target.options.replyMarkup
	p.MessageEffectID = target.options.messageEffectID
	p.ReceiverUserID = target.receiverID
	p.CallbackQueryID = target.callbackID
	p.ReplyParameters = target.reply
}

// Document sends a document by file ID or URL and discards the returned message.
func (c *Context) Document(file, caption string, options ...SendOption) error {
	_, err := c.DocumentMessage(file, caption, options...)
	return err
}

// DocumentMessage sends a document and returns Telegram's message.
func (c *Context) DocumentMessage(file, caption string, options ...SendOption) (*Message, error) {
	target, err := c.mediaTarget(false, options)
	if err != nil {
		return nil, err
	}
	p := SendDocumentParams{Document: file, Caption: caption}
	applyDocumentOptions(&p, target)
	return c.Bot.SendDocument(c.Context, p)
}

// EphemeralDocument sends a document visible only to the current sender.
func (c *Context) EphemeralDocument(file, caption string, options ...SendOption) error {
	target, err := c.mediaTarget(true, options)
	if err != nil {
		return err
	}
	p := SendDocumentParams{Document: file, Caption: caption}
	applyDocumentOptions(&p, target)
	_, err = c.Bot.SendDocument(c.Context, p)
	return err
}

func applyVideoOptions(p *SendVideoParams, target contextSendTarget) {
	p.ChatID = target.chatID
	p.ParseMode = target.options.parseMode
	p.DisableNotification = target.options.disableNotification
	p.ProtectContent = target.options.protectContent
	p.MessageThreadID = target.options.messageThreadID
	p.ReplyMarkup = target.options.replyMarkup
	p.MessageEffectID = target.options.messageEffectID
	p.HasSpoiler = target.options.hasSpoiler
	p.ShowCaptionAboveMedia = target.options.showCaptionAbove
	p.SupportsStreaming = target.options.supportsStreaming
	p.ReceiverUserID = target.receiverID
	p.CallbackQueryID = target.callbackID
	p.ReplyParameters = target.reply
}

// Video sends a video by file ID or URL.
func (c *Context) Video(file, caption string, options ...SendOption) error {
	target, err := c.mediaTarget(false, options)
	if err != nil {
		return err
	}
	p := SendVideoParams{Video: file, Caption: caption}
	applyVideoOptions(&p, target)
	_, err = c.Bot.SendVideo(c.Context, p)
	return err
}

// EphemeralVideo sends a video visible only to the current sender.
func (c *Context) EphemeralVideo(file, caption string, options ...SendOption) error {
	target, err := c.mediaTarget(true, options)
	if err != nil {
		return err
	}
	p := SendVideoParams{Video: file, Caption: caption}
	applyVideoOptions(&p, target)
	_, err = c.Bot.SendVideo(c.Context, p)
	return err
}

func applyAnimationOptions(p *SendAnimationParams, target contextSendTarget) {
	p.ChatID = target.chatID
	p.ParseMode = target.options.parseMode
	p.DisableNotification = target.options.disableNotification
	p.ProtectContent = target.options.protectContent
	p.MessageThreadID = target.options.messageThreadID
	p.ReplyMarkup = target.options.replyMarkup
	p.MessageEffectID = target.options.messageEffectID
	p.HasSpoiler = target.options.hasSpoiler
	p.ShowCaptionAboveMedia = target.options.showCaptionAbove
	p.ReceiverUserID = target.receiverID
	p.CallbackQueryID = target.callbackID
	p.ReplyParameters = target.reply
}

// Animation sends an animation by file ID or URL.
func (c *Context) Animation(file, caption string, options ...SendOption) error {
	target, err := c.mediaTarget(false, options)
	if err != nil {
		return err
	}
	p := SendAnimationParams{Animation: file, Caption: caption}
	applyAnimationOptions(&p, target)
	_, err = c.Bot.SendAnimation(c.Context, p)
	return err
}

// EphemeralAnimation sends an animation visible only to the current sender.
func (c *Context) EphemeralAnimation(file, caption string, options ...SendOption) error {
	target, err := c.mediaTarget(true, options)
	if err != nil {
		return err
	}
	p := SendAnimationParams{Animation: file, Caption: caption}
	applyAnimationOptions(&p, target)
	_, err = c.Bot.SendAnimation(c.Context, p)
	return err
}

func applyAudioOptions(p *SendAudioParams, target contextSendTarget) {
	p.ChatID = target.chatID
	p.ParseMode = target.options.parseMode
	p.DisableNotification = target.options.disableNotification
	p.ProtectContent = target.options.protectContent
	p.MessageThreadID = target.options.messageThreadID
	p.ReplyMarkup = target.options.replyMarkup
	p.MessageEffectID = target.options.messageEffectID
	p.ReceiverUserID = target.receiverID
	p.CallbackQueryID = target.callbackID
	p.ReplyParameters = target.reply
}

// Audio sends an audio file by file ID or URL.
func (c *Context) Audio(file, caption string, options ...SendOption) error {
	target, err := c.mediaTarget(false, options)
	if err != nil {
		return err
	}
	p := SendAudioParams{Audio: file, Caption: caption}
	applyAudioOptions(&p, target)
	_, err = c.Bot.SendAudio(c.Context, p)
	return err
}

// EphemeralAudio sends audio visible only to the current sender.
func (c *Context) EphemeralAudio(file, caption string, options ...SendOption) error {
	target, err := c.mediaTarget(true, options)
	if err != nil {
		return err
	}
	p := SendAudioParams{Audio: file, Caption: caption}
	applyAudioOptions(&p, target)
	_, err = c.Bot.SendAudio(c.Context, p)
	return err
}

func applyVoiceOptions(p *SendVoiceParams, target contextSendTarget) {
	p.ChatID = target.chatID
	p.ParseMode = target.options.parseMode
	p.DisableNotification = target.options.disableNotification
	p.ProtectContent = target.options.protectContent
	p.MessageThreadID = target.options.messageThreadID
	p.ReplyMarkup = target.options.replyMarkup
	p.MessageEffectID = target.options.messageEffectID
	p.ReceiverUserID = target.receiverID
	p.CallbackQueryID = target.callbackID
	p.ReplyParameters = target.reply
}

// Voice sends a voice note by file ID or URL.
func (c *Context) Voice(file, caption string, options ...SendOption) error {
	target, err := c.mediaTarget(false, options)
	if err != nil {
		return err
	}
	p := SendVoiceParams{Voice: file, Caption: caption}
	applyVoiceOptions(&p, target)
	_, err = c.Bot.SendVoice(c.Context, p)
	return err
}

// EphemeralVoice sends a voice note visible only to the current sender.
func (c *Context) EphemeralVoice(file, caption string, options ...SendOption) error {
	target, err := c.mediaTarget(true, options)
	if err != nil {
		return err
	}
	p := SendVoiceParams{Voice: file, Caption: caption}
	applyVoiceOptions(&p, target)
	_, err = c.Bot.SendVoice(c.Context, p)
	return err
}

func applyStickerOptions(p *SendStickerParams, target contextSendTarget) {
	p.ChatID = target.chatID
	p.DisableNotification = target.options.disableNotification
	p.ProtectContent = target.options.protectContent
	p.MessageThreadID = target.options.messageThreadID
	p.ReplyMarkup = target.options.replyMarkup
	p.MessageEffectID = target.options.messageEffectID
	p.ReceiverUserID = target.receiverID
	p.CallbackQueryID = target.callbackID
	p.ReplyParameters = target.reply
}

// Sticker sends a sticker by file ID or URL.
func (c *Context) Sticker(file string, options ...SendOption) error {
	target, err := c.mediaTarget(false, options)
	if err != nil {
		return err
	}
	p := SendStickerParams{Sticker: file}
	applyStickerOptions(&p, target)
	_, err = c.Bot.SendSticker(c.Context, p)
	return err
}

// EphemeralSticker sends a sticker visible only to the current sender.
func (c *Context) EphemeralSticker(file string, options ...SendOption) error {
	target, err := c.mediaTarget(true, options)
	if err != nil {
		return err
	}
	p := SendStickerParams{Sticker: file}
	applyStickerOptions(&p, target)
	_, err = c.Bot.SendSticker(c.Context, p)
	return err
}
