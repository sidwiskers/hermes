package framework

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/sidwiskers/hermes/api"
)

var (
	// ErrWebhookMethod reports an invalid direct webhook-response method.
	ErrWebhookMethod = errors.New("hermes: invalid webhook response method")
	// ErrWebhookResponseSet reports a second response from the same handler.
	ErrWebhookResponseSet = errors.New("hermes: webhook response is already set")
)

// WebhookResponse is an optional Bot API call returned directly from a
// synchronous webhook request. Params should be the typed parameter struct for
// Method.
type WebhookResponse struct {
	Method string
	Params any
}

type contextState struct {
	webhook    WebhookResponse
	webhookSet bool
}

// Context is the per-update handler view. Handler contexts may be pooled;
// call Clone before retaining one after the handler returns.
//
// Bot is the low-level typed client. Message and Callback are derived from
// Update for convenient access.
type Context struct {
	context.Context
	Bot      *api.Client
	Update   *Update
	Message  *Message
	Callback *CallbackQuery

	command string
	args    string
	state   contextState
	shared  *contextState
}

// NewContext builds a context for an independent Router. username is used to
// reject commands addressed to another bot and should be lowercase without a
// leading @.
func NewContext(ctx context.Context, bot *api.Client, update *Update, username string) *Context {
	c := new(Context)
	c.reset(ctx, bot, update, username)
	return c
}

func (c *Context) reset(ctx context.Context, bot *api.Client, update *Update, username string) {
	*c = Context{Context: ctx, Bot: bot, Update: update}
	c.shared = &c.state
	if update == nil {
		return
	}
	c.Message = update.PrimaryMessage()
	c.Callback = update.CallbackQuery
	if update.CallbackQuery == nil {
		c.command, c.args = parseCommand(c.Message, username)
	}
}

// Clone returns an independent Context value that may be retained after the
// current handler returns. Ordinary handler contexts are borrowed and may be
// reused when context pooling is enabled.
func (c *Context) Clone() *Context {
	if c == nil {
		return nil
	}
	cloned := *c
	if c.shared != nil {
		cloned.state = *c.shared
	}
	cloned.shared = &cloned.state
	return &cloned
}

// Command returns the normalized command name without a slash or bot mention.
func (c *Context) Command() string {
	if c == nil {
		return ""
	}
	return c.command
}

// Args returns the trimmed text after the command token.
func (c *Context) Args() string {
	if c == nil {
		return ""
	}
	return c.args
}

// Argv splits Args on Unicode whitespace.
func (c *Context) Argv() []string { return strings.Fields(c.Args()) }

// Type reports the concrete update type.
func (c *Context) Type() UpdateType {
	if c == nil || c.Update == nil {
		return UpdateUnknown
	}
	return c.Update.Type()
}

// Sender returns the user responsible for the update when Telegram supplies one.
func (c *Context) Sender() *User {
	if c == nil || c.Update == nil {
		return nil
	}
	return c.Update.Sender()
}

// Text returns the primary message text or caption.
func (c *Context) Text() string {
	if c == nil || c.Message == nil {
		return ""
	}
	return c.Message.ContentText()
}

// Data returns callback-query data.
func (c *Context) Data() string {
	if c == nil || c.Callback == nil {
		return ""
	}
	return c.Callback.Data
}

// Chat returns the primary message chat.
func (c *Context) Chat() *Chat {
	if c == nil || c.Message == nil {
		return nil
	}
	return &c.Message.Chat
}

// ChatID returns the primary message chat ID and whether it is available.
func (c *Context) ChatID() (int64, bool) {
	if c == nil || c.Message == nil {
		return 0, false
	}
	return c.Message.Chat.ID, c.Message.Chat.ID != 0
}

// MessageID returns the ordinary primary message ID and whether it is available.
func (c *Context) MessageID() (int, bool) {
	if c == nil || c.Message == nil || c.Message.MessageID == 0 {
		return 0, false
	}
	return c.Message.MessageID, true
}

// RespondWebhook configures a Bot API call to be returned in the current
// synchronous webhook HTTP response. Use Bot.WebhookReplyHandler; the ordinary
// queued WebhookHandler acknowledges before handlers run and ignores it.
// Exactly one response may be configured per update.
func (c *Context) RespondWebhook(method string, params any) error {
	if c == nil || !validWebhookMethod(method) {
		return ErrWebhookMethod
	}
	state := c.sharedState()
	if state.webhookSet {
		return ErrWebhookResponseSet
	}
	state.webhook = WebhookResponse{Method: method, Params: params}
	state.webhookSet = true
	return nil
}

// DirectWebhookResponse returns the response configured by RespondWebhook.
// It is primarily useful to custom synchronous webhook integrations.
func (c *Context) DirectWebhookResponse() (WebhookResponse, bool) {
	if c == nil {
		return WebhookResponse{}, false
	}
	state := c.sharedState()
	return state.webhook, state.webhookSet
}

func (c *Context) sharedState() *contextState {
	if c.shared == nil {
		c.shared = &c.state
	}
	return c.shared
}

func validWebhookMethod(method string) bool {
	if method == "" || len(method) > 128 {
		return false
	}
	for index := 0; index < len(method); index++ {
		char := method[index]
		switch {
		case char >= 'a' && char <= 'z':
		case char >= 'A' && char <= 'Z':
		case char >= '0' && char <= '9':
		case char == '_':
		default:
			return false
		}
	}
	return true
}

type sendOptions struct {
	parseMode           string
	disableNotification bool
	protectContent      bool
	noPreview           bool
	messageThreadID     int
	allowWithoutReply   bool
	replyMarkup         ReplyMarkup
	hasSpoiler          bool
	showCaptionAbove    bool
	supportsStreaming   bool
	messageEffectID     string
}

// SendOption configures concise Context send helpers.
type SendOption func(*sendOptions)

var (
	// HTML selects Telegram HTML parsing.
	HTML SendOption = func(p *sendOptions) { p.parseMode = ParseModeHTML }
	// Markdown selects Telegram's legacy Markdown parsing.
	Markdown SendOption = func(p *sendOptions) { p.parseMode = ParseModeMarkdown }
	// MarkdownV2 selects Telegram MarkdownV2 parsing.
	MarkdownV2 SendOption = func(p *sendOptions) { p.parseMode = ParseModeMarkdownV2 }
	// Silent disables message notification.
	Silent SendOption = func(p *sendOptions) { p.disableNotification = true }
	// Protected prevents forwarding and saving.
	Protected SendOption = func(p *sendOptions) { p.protectContent = true }
	// NoPreview disables link previews.
	NoPreview SendOption = func(p *sendOptions) { p.noPreview = true }
	// Spoiler covers supported media with a spoiler animation.
	Spoiler SendOption = func(p *sendOptions) { p.hasSpoiler = true }
	// CaptionAbove places a media caption above its media.
	CaptionAbove SendOption = func(p *sendOptions) { p.showCaptionAbove = true }
	// Streaming marks supported video as suitable for streaming.
	Streaming SendOption = func(p *sendOptions) { p.supportsStreaming = true }
)

// WithKeyboard attaches an inline keyboard.
func WithKeyboard(keyboard InlineKeyboardMarkup) SendOption { return WithMarkup(keyboard) }

// WithMarkup attaches any supported inline or reply markup.
func WithMarkup(markup ReplyMarkup) SendOption {
	return func(p *sendOptions) { p.replyMarkup = markup }
}

// InThread sends to a forum topic or message thread.
func InThread(threadID int) SendOption { return func(p *sendOptions) { p.messageThreadID = threadID } }

// AllowWithoutReply permits a reply even if its target no longer exists.
func AllowWithoutReply() SendOption { return func(p *sendOptions) { p.allowWithoutReply = true } }

// WithEffect attaches a Telegram message-effect identifier.
func WithEffect(effectID string) SendOption {
	return func(p *sendOptions) { p.messageEffectID = effectID }
}

func resolveSendOptions(options []SendOption) sendOptions {
	var result sendOptions
	for _, option := range options {
		if option != nil {
			option(&result)
		}
	}
	return result
}

// Send sends text to the current chat and discards the returned message.
func (c *Context) Send(text string, options ...SendOption) error {
	_, err := c.SendMessage(text, options...)
	return err
}

// SendMessage sends text to the current chat and returns Telegram's message.
func (c *Context) SendMessage(text string, options ...SendOption) (*Message, error) {
	chatID, ok := c.ChatID()
	if !ok {
		return nil, fmt.Errorf("hermes: update has no chat")
	}
	params := SendMessageParams{ChatID: chatID, Text: text}
	applyTextOptions(&params, resolveSendOptions(options))
	return c.Bot.SendMessage(c.Context, params)
}

// Reply replies to the current message and discards the returned message.
func (c *Context) Reply(text string, options ...SendOption) error {
	_, err := c.ReplyMessage(text, options...)
	return err
}

// ReplyMessage replies to the current ordinary or ephemeral message.
func (c *Context) ReplyMessage(text string, options ...SendOption) (*Message, error) {
	chatID, ok := c.ChatID()
	if !ok || c.Message == nil {
		return nil, fmt.Errorf("hermes: update has no replyable message")
	}
	params := SendMessageParams{ChatID: chatID, Text: text, ReplyParameters: c.replyParameters()}
	applyTextOptions(&params, resolveSendOptions(options))
	return c.Bot.SendMessage(c.Context, params)
}

// Ephemeral sends text visible only to the current sender.
func (c *Context) Ephemeral(text string, options ...SendOption) error {
	_, err := c.EphemeralMessage(text, options...)
	return err
}

// EphemeralMessage sends private text and returns Telegram's message.
func (c *Context) EphemeralMessage(text string, options ...SendOption) (*Message, error) {
	chatID, ok := c.ChatID()
	if !ok {
		return nil, fmt.Errorf("hermes: update has no chat")
	}
	sender := c.Sender()
	if sender == nil {
		return nil, fmt.Errorf("hermes: update has no sender")
	}
	params := SendMessageParams{ChatID: chatID, Text: text, ReceiverUserID: sender.ID}
	c.applyEphemeralTarget(&params.CallbackQueryID, &params.ReplyParameters)
	applyTextOptions(&params, resolveSendOptions(options))
	return c.Bot.SendMessage(c.Context, params)
}

// Answer responds to the current callback query with a notification.
func (c *Context) Answer(text string) error {
	if c == nil || c.Callback == nil {
		return fmt.Errorf("hermes: update has no callback query")
	}
	return c.Bot.AnswerCallback(c.Context, AnswerCallbackQueryParams{CallbackQueryID: c.Callback.ID, Text: text})
}

// Acknowledge silently answers the current callback query.
func (c *Context) Acknowledge() error { return c.Answer("") }

// Alert responds to the current callback query with an alert dialog.
func (c *Context) Alert(text string) error {
	if c == nil || c.Callback == nil {
		return fmt.Errorf("hermes: update has no callback query")
	}
	return c.Bot.AnswerCallback(c.Context, AnswerCallbackQueryParams{CallbackQueryID: c.Callback.ID, Text: text, ShowAlert: true})
}

// ChatAction sends an activity indicator, such as typing, to the current chat.
func (c *Context) ChatAction(action string) error {
	chatID, ok := c.ChatID()
	if !ok {
		return fmt.Errorf("hermes: update has no chat")
	}
	threadID := 0
	if c.Message != nil {
		threadID = c.Message.MessageThreadID
	}
	return c.Bot.SendChatAction(c.Context, SendChatActionParams{ChatID: chatID, MessageThreadID: threadID, Action: action})
}

func (c *Context) replyParameters() *ReplyParameters {
	if c == nil || c.Message == nil {
		return nil
	}
	if c.Message.EphemeralMessageID != 0 {
		return &ReplyParameters{EphemeralMessageID: c.Message.EphemeralMessageID}
	}
	return &ReplyParameters{MessageID: c.Message.MessageID}
}

func (c *Context) applyEphemeralTarget(callbackID *string, reply **ReplyParameters) {
	switch {
	case c != nil && c.Callback != nil:
		*callbackID = c.Callback.ID
	case c != nil && c.Message != nil && c.Message.EphemeralMessageID != 0:
		*reply = &ReplyParameters{EphemeralMessageID: c.Message.EphemeralMessageID}
	}
}

func applyTextOptions(params *SendMessageParams, options sendOptions) {
	params.ParseMode = options.parseMode
	params.DisableNotification = options.disableNotification
	params.ProtectContent = options.protectContent
	params.MessageThreadID = options.messageThreadID
	params.ReplyMarkup = options.replyMarkup
	params.MessageEffectID = options.messageEffectID
	if options.noPreview {
		params.LinkPreviewOptions = &LinkPreviewOptions{IsDisabled: true}
	}
	if options.allowWithoutReply {
		if params.ReplyParameters == nil {
			params.ReplyParameters = &ReplyParameters{}
		}
		params.ReplyParameters.AllowSendingWithoutReply = true
	}
}

func parseCommand(message *Message, botUsername string) (string, string) {
	if message == nil || !strings.HasPrefix(message.Text, "/") {
		return "", ""
	}
	text := strings.TrimSpace(message.Text)
	if text == "" {
		return "", ""
	}
	token, args := text, ""
	if index := strings.IndexFunc(text, unicode.IsSpace); index >= 0 {
		token, args = text[:index], strings.TrimSpace(text[index:])
	}
	token = strings.TrimPrefix(token, "/")
	name, mention, hasMention := strings.Cut(token, "@")
	if name == "" {
		return "", ""
	}
	if hasMention && (botUsername == "" || !strings.EqualFold(mention, botUsername)) {
		return "", ""
	}
	return strings.ToLower(name), args
}
