package hermes

import (
	"context"
	"log/slog"
	"time"

	"github.com/sidwiskers/hermes/framework"
)

// Framework types are re-exported through the compact root package.
type (
	Context         = framework.Context
	Handler         = framework.Handler
	Middleware      = framework.Middleware
	Filter          = framework.Filter
	Router          = framework.Router
	Group           = framework.Group
	SendOption      = framework.SendOption
	PanicError      = framework.PanicError
	WebhookResponse = framework.WebhookResponse
)

var (
	// HTML, Markdown, and MarkdownV2 select Telegram text parsing modes.
	HTML       = framework.HTML
	Markdown   = framework.Markdown
	MarkdownV2 = framework.MarkdownV2
	// Silent disables notification; Protected disables forwarding and saving;
	// NoPreview disables link previews.
	Silent    = framework.Silent
	Protected = framework.Protected
	NoPreview = framework.NoPreview
	// Spoiler, CaptionAbove, and Streaming tune media delivery.
	Spoiler      = framework.Spoiler
	CaptionAbove = framework.CaptionAbove
	Streaming    = framework.Streaming
	// Direct webhook response validation errors.
	ErrWebhookMethod      = framework.ErrWebhookMethod
	ErrWebhookResponseSet = framework.ErrWebhookResponseSet
)

// NewRouter creates an independent concurrency-safe router.
func NewRouter() *Router { return framework.NewRouter() }

func newContext(ctx context.Context, bot *Bot, update *Update) *Context {
	if bot == nil {
		return framework.NewContext(ctx, nil, update, "")
	}
	return framework.NewContext(ctx, bot.Client, update, bot.loadUsername())
}

// WithKeyboard attaches an inline keyboard to a context send helper.
func WithKeyboard(keyboard InlineKeyboardMarkup) SendOption { return framework.WithKeyboard(keyboard) }

// WithMarkup attaches any supported inline or reply markup.
func WithMarkup(markup ReplyMarkup) SendOption { return framework.WithMarkup(markup) }

// InThread sends to a forum topic or message thread.
func InThread(threadID int) SendOption { return framework.InThread(threadID) }

// AllowWithoutReply permits a reply even if its target no longer exists.
func AllowWithoutReply() SendOption { return framework.AllowWithoutReply() }

// WithEffect attaches a Telegram message-effect identifier.
func WithEffect(effectID string) SendOption { return framework.WithEffect(effectID) }

// All matches when every non-nil filter matches.
func All(filters ...Filter) Filter { return framework.All(filters...) }

// Any matches when at least one non-nil filter matches.
func Any(filters ...Filter) Filter { return framework.Any(filters...) }

// Not negates a filter. A nil filter is treated as false before negation.
func Not(filter Filter) Filter { return framework.Not(filter) }

// UpdateIs matches any of the supplied update types.
func UpdateIs(types ...UpdateType) Filter { return framework.UpdateIs(types...) }

// MessageUpdate matches updates with a primary message.
func MessageUpdate(c *Context) bool { return framework.MessageUpdate(c) }

// CallbackUpdate matches callback-query updates.
func CallbackUpdate(c *Context) bool { return framework.CallbackUpdate(c) }

// TextMessage matches messages containing text.
func TextMessage(c *Context) bool { return framework.TextMessage(c) }

// CaptionedMessage matches messages containing a caption.
func CaptionedMessage(c *Context) bool { return framework.CaptionedMessage(c) }

// PhotoMessage matches messages containing photos.
func PhotoMessage(c *Context) bool { return framework.PhotoMessage(c) }

// DocumentMessage matches messages containing a document.
func DocumentMessage(c *Context) bool { return framework.DocumentMessage(c) }

// StickerMessage matches messages containing a sticker.
func StickerMessage(c *Context) bool { return framework.StickerMessage(c) }

// VideoMessage matches messages containing a video.
func VideoMessage(c *Context) bool { return framework.VideoMessage(c) }

// VoiceMessage matches messages containing a voice note.
func VoiceMessage(c *Context) bool { return framework.VoiceMessage(c) }

// PrivateChat matches messages from private chats.
func PrivateChat(c *Context) bool { return framework.PrivateChat(c) }

// GroupChat matches messages from groups and supergroups.
func GroupChat(c *Context) bool { return framework.GroupChat(c) }

// ChannelChat matches channel messages.
func ChannelChat(c *Context) bool { return framework.ChannelChat(c) }

// EphemeralMessage matches Telegram ephemeral messages.
func EphemeralMessage(c *Context) bool { return framework.EphemeralMessage(c) }

// FromUsers matches updates sent by any supplied user ID.
func FromUsers(ids ...int64) Filter { return framework.FromUsers(ids...) }

// InChats matches updates belonging to any supplied chat ID.
func InChats(ids ...int64) Filter { return framework.InChats(ids...) }

// TextEquals matches message text or captions exactly.
func TextEquals(values ...string) Filter { return framework.TextEquals(values...) }

// TextPrefix matches message text or captions by prefix.
func TextPrefix(prefixes ...string) Filter { return framework.TextPrefix(prefixes...) }

// CallbackDataPrefix matches callback data by prefix.
func CallbackDataPrefix(prefixes ...string) Filter { return framework.CallbackDataPrefix(prefixes...) }

// Recover converts downstream panics to PanicError values.
func Recover() Middleware { return framework.Recover() }

// Timeout gives each downstream handler a derived deadline.
func Timeout(duration time.Duration) Middleware { return framework.Timeout(duration) }

// RecoverWith is Recover with an additional panic-report callback.
func RecoverWith(report func(*Context, *PanicError)) Middleware { return framework.RecoverWith(report) }

// Logger emits one structured slog record after each routed update.
func Logger(logger *slog.Logger) Middleware { return framework.Logger(logger) }
