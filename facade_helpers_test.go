package hermes

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestFacadeConstructorsAndOptions(t *testing.T) {
	t.Parallel()

	if encoded, err := MultipartJSON(map[string]int{"value": 1}); err != nil || encoded != `{"value":1}` {
		t.Fatalf("MultipartJSON() = %q, %v", encoded, err)
	}
	if MultipartInt(42) != "42" || Attachment("photo") != "attach://photo" {
		t.Fatal("multipart facade helpers returned unexpected values")
	}
	if value := Bool(false); value == nil || *value {
		t.Fatalf("Bool(false) = %#v", value)
	}
	if upload := NewUpload("photo", "photo.jpg", strings.NewReader("image")); upload.Field != "photo" || upload.Name != "photo.jpg" {
		t.Fatalf("upload = %#v", upload)
	}

	if DefaultCommandScope().Type != "default" || AllPrivateChatsCommandScope().Type != "all_private_chats" ||
		AllGroupChatsCommandScope().Type != "all_group_chats" || AllChatAdministratorsCommandScope().Type != "all_chat_administrators" {
		t.Fatal("command scope facade helpers returned unexpected types")
	}
	if ChatCommandScope(1).ChatID != 1 || ChatAdministratorsCommandScope(2).ChatID != 2 {
		t.Fatal("chat command scopes lost chat IDs")
	}
	if scope := ChatMemberCommandScope(3, 4); scope.ChatID != 3 || scope.UserID != 4 {
		t.Fatalf("member scope = %#v", scope)
	}
	if CommandsMenuButton().Type != MenuButtonTypeCommands || DefaultMenuButton().Type != MenuButtonTypeDefault ||
		WebAppMenuButton("Open", "https://example.com").Type != MenuButtonTypeWebApp {
		t.Fatal("menu button facade helpers returned unexpected types")
	}

	keyboard := Keyboard(Row(Button("Data", "data"), URLButton("URL", "https://example.com"),
		WebAppButton("App", "https://example.com"), CopyButton("Copy", "value"), PayButton("Pay")))
	if len(keyboard.InlineKeyboard) != 1 || len(keyboard.InlineKeyboard[0]) != 5 {
		t.Fatalf("keyboard = %#v", keyboard)
	}
	reply := ReplyKeyboard(KeyRow(Key("One")))
	_ = RemoveKeyboard(true)
	_ = NewForceReply("Reply")
	_ = WithKeyboard(keyboard)
	_ = WithMarkup(reply)
	_ = InThread(7)
	_ = AllowWithoutReply()
	_ = WithEffect("effect")

	if EmojiReaction("🔥").Type != ReactionEmoji || CustomEmojiReaction("emoji").Type != ReactionCustomEmoji || PaidReaction().Type != ReactionPaid {
		t.Fatal("reaction facade helpers returned unexpected types")
	}

	update, err := DecodeUpdate([]byte(`{"update_id":1,"message":{"message_id":2,"chat":{"id":3,"type":"private"},"date":1,"text":"hello"}}`), false)
	if err != nil {
		t.Fatal(err)
	}
	updates, err := DecodeUpdates([]byte(`[{"update_id":1}]`), true)
	if err != nil || len(updates) != 1 {
		t.Fatalf("DecodeUpdates() = %#v, %v", updates, err)
	}

	ctx := newContext(context.Background(), nil, &update)
	filters := []Filter{
		All(MessageUpdate, TextMessage), Any(TextMessage, CallbackUpdate), Not(CallbackUpdate),
		UpdateIs(UpdateMessage), FromUsers(1), InChats(3), TextEquals("hello"),
		TextPrefix("he"), CallbackDataPrefix("callback:"),
	}
	for _, filter := range filters {
		_ = filter(ctx)
	}
	_ = CaptionedMessage(ctx)
	_ = PhotoMessage(ctx)
	_ = DocumentMessage(ctx)
	_ = StickerMessage(ctx)
	_ = VideoMessage(ctx)
	_ = VoiceMessage(ctx)
	_ = PrivateChat(ctx)
	_ = GroupChat(ctx)
	_ = ChannelChat(ctx)
	_ = EphemeralMessage(ctx)
	_ = NewRouter()
	_ = Recover()
	_ = Timeout(time.Second)
	_ = RecoverWith(func(*Context, *PanicError) {})
	_ = Logger(slog.Default())

	var handled error
	bot := New("TOKEN",
		WithUserAgent("hermes-test"),
		WithResponseLimit(1024),
		WithRawUpdates(true),
		WithTestEnvironment(true),
		WithContextPooling(false),
		WithErrorHandler(func(_ *Context, err error) { handled = err }),
	)
	want := errors.New("handled")
	bot.report(newContext(context.Background(), bot, &update), want)
	if !errors.Is(handled, want) {
		t.Fatalf("handled error = %v", handled)
	}
	bot.SetErrorHandler(func(_ *Context, err error) { handled = err })
	if _, err := Call[int](context.Background(), nil, "method", nil); !errors.Is(err, ErrClientRequired) {
		t.Fatalf("nil Bot Call error = %v", err)
	}
}
