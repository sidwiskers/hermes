//go:build integration

// Package integration contains opt-in tests against Telegram's Bot API. The
// separate test environment is the default; HERMES_TEST_PRODUCTION=true opts
// into the production endpoint for a dedicated disposable bot. These tests
// create and delete messages and may change webhook state.
package integration

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sidwiskers/hermes"
)

func liveBot(t *testing.T) (*hermes.Bot, any) {
	t.Helper()
	token := strings.TrimSpace(os.Getenv("HERMES_TEST_BOT_TOKEN"))
	chat := strings.TrimSpace(os.Getenv("HERMES_TEST_CHAT_ID"))
	if token == "" || chat == "" {
		t.Skip("HERMES_TEST_BOT_TOKEN and HERMES_TEST_CHAT_ID are required")
	}
	var chatID any = chat
	if numeric, err := strconv.ParseInt(chat, 10, 64); err == nil {
		chatID = numeric
	}
	return hermes.New(token, liveEndpointOptions(t)...), chatID
}

func liveEndpointOptions(t *testing.T) []hermes.Option {
	t.Helper()
	if !liveBool(t, "HERMES_TEST_PRODUCTION") {
		return []hermes.Option{hermes.WithTestEnvironment(true)}
	}
	return nil
}

func liveBool(t *testing.T, name string) bool {
	t.Helper()
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return false
	}
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		t.Fatalf("%s must be a boolean: %v", name, err)
	}
	return enabled
}

func cleanupMessage(t *testing.T, bot *hermes.Bot, chatID any, messageID int) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = bot.DeleteMessage(ctx, hermes.DeleteMessageParams{
			ChatID: chatID, MessageID: messageID,
		})
	})
}

func TestLiveIdentity(t *testing.T) {
	bot, _ := liveBot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	me, err := bot.GetMe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !me.IsBot || me.ID == 0 || me.Username == "" {
		t.Fatalf("unexpected bot identity: %#v", me)
	}
	if expected := strings.TrimPrefix(strings.TrimSpace(os.Getenv("HERMES_TEST_BOT_USERNAME")), "@"); expected != "" && !strings.EqualFold(me.Username, expected) {
		t.Fatalf("bot username = %q, want %q", me.Username, expected)
	}
}

func TestLiveMessageLifecycle(t *testing.T) {
	bot, chatID := liveBot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	message, err := bot.SendMessage(ctx, hermes.SendMessageParams{
		ChatID: chatID,
		Text:   "Hermes live conformance",
	})
	if err != nil {
		t.Fatal(err)
	}
	cleanupMessage(t, bot, chatID, message.MessageID)

	edited, err := bot.EditMessageText(ctx, hermes.EditMessageTextParams{
		ChatID: chatID, MessageID: message.MessageID, Text: "Hermes live conformance: edited",
	})
	if err != nil {
		t.Fatal(err)
	}
	if edited == nil || edited.MessageID != message.MessageID {
		t.Fatalf("unexpected edited message: %#v", edited)
	}
	if err := bot.DeleteMessage(ctx, hermes.DeleteMessageParams{
		ChatID: chatID, MessageID: message.MessageID,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestLiveStreamedUpload(t *testing.T) {
	bot, chatID := liveBot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	const content = "Hermes Bot API conformance\n"
	message, err := bot.SendDocumentUpload(ctx, hermes.SendDocumentParams{
		ChatID: chatID, Caption: "Hermes streamed upload",
	}, "conformance.txt", strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	cleanupMessage(t, bot, chatID, message.MessageID)
	if message.Document == nil || message.Document.FileID == "" {
		t.Fatalf("unexpected uploaded document: %#v", message.Document)
	}
	file, err := bot.GetFile(ctx, message.Document.FileID)
	if err != nil {
		t.Fatal(err)
	}
	if file.FilePath == "" {
		t.Fatalf("uploaded file has no download path: %#v", file)
	}
	var downloaded strings.Builder
	written, err := bot.DownloadFile(ctx, file.FilePath, &downloaded)
	if err != nil {
		t.Fatal(err)
	}
	if written != int64(len(content)) || downloaded.String() != content {
		t.Fatalf("downloaded %d bytes %q, want %d bytes %q", written, downloaded.String(), len(content), content)
	}
	if err := bot.DeleteMessage(ctx, hermes.DeleteMessageParams{
		ChatID: chatID, MessageID: message.MessageID,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestLivePollingDelivery(t *testing.T) {
	expected := strings.TrimSpace(os.Getenv("HERMES_TEST_EXPECT_TEXT"))
	if expected == "" {
		t.Skip("HERMES_TEST_EXPECT_TEXT is required for an inbound polling probe")
	}
	bot, _ := liveBot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	received := make(chan int64, 1)
	bot.On(hermes.TextMessage, func(c *hermes.Context) error {
		if c.Text() != expected {
			return nil
		}
		select {
		case received <- c.Update.UpdateID:
			cancel()
		default:
		}
		return nil
	})
	if err := bot.Poll(ctx, hermes.PollOptions{
		Timeout: 1, Limit: 100, AllowedUpdates: []string{"message"},
	}); err != nil {
		t.Fatal(err)
	}
	select {
	case updateID := <-received:
		if updateID == 0 {
			t.Fatal("received update has no identifier")
		}
	default:
		t.Fatalf("did not receive %q before polling stopped", expected)
	}
}

func TestLivePollingCancellation(t *testing.T) {
	bot, _ := liveBot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	if err := bot.Poll(ctx, hermes.PollOptions{Timeout: 1, Limit: 1}); err != nil {
		t.Fatal(err)
	}
}

func TestLiveFloodWait(t *testing.T) {
	if !liveBool(t, "HERMES_TEST_FLOOD_WAIT") {
		t.Skip("set HERMES_TEST_FLOOD_WAIT=true to enable the bounded live rate-limit probe")
	}
	bot, chatID := liveBot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	const attempts = 64
	results := make(chan error, attempts)
	start := make(chan struct{})
	for range attempts {
		go func() {
			<-start
			results <- bot.SendChatAction(ctx, hermes.SendChatActionParams{
				ChatID: chatID,
				Action: "typing",
			})
		}()
	}
	close(start)

	accepted := 0
	retryAfter := 0
	var unexpected []error
	for range attempts {
		err := <-results
		if err == nil {
			accepted++
			continue
		}
		var apiErr *hermes.APIError
		if errors.As(err, &apiErr) && apiErr.Code == 429 && apiErr.RetryAfter() > 0 {
			if apiErr.RetryAfter() > retryAfter {
				retryAfter = apiErr.RetryAfter()
			}
			continue
		}
		unexpected = append(unexpected, err)
	}

	t.Logf("bounded flood-wait report: attempts=%d accepted=%d limited=%d retry_after=%d", attempts, accepted, attempts-accepted-len(unexpected), retryAfter)
	if len(unexpected) != 0 {
		t.Fatalf("live rate-limit probe returned %d unexpected errors: %v", len(unexpected), unexpected[0])
	}
	if retryAfter <= 0 {
		t.Fatal("live rate-limit probe did not observe a 429 response with retry_after")
	}
}

func TestLiveWebhookConfiguration(t *testing.T) {
	bot, _ := liveBot(t)
	webhookURL := strings.TrimSpace(os.Getenv("HERMES_TEST_WEBHOOK_URL"))
	secret := strings.TrimSpace(os.Getenv("HERMES_TEST_WEBHOOK_SECRET"))
	if webhookURL == "" || secret == "" {
		t.Skip("HERMES_TEST_WEBHOOK_URL and HERMES_TEST_WEBHOOK_SECRET are required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	keepWebhook := liveBool(t, "HERMES_TEST_KEEP_WEBHOOK")
	if !keepWebhook {
		t.Cleanup(func() {
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cleanupCancel()
			_ = bot.DeleteWebhook(cleanupCtx, hermes.DeleteWebhookParams{})
		})
	}

	if err := bot.SetWebhook(ctx, hermes.SetWebhookParams{
		URL: webhookURL, SecretToken: secret,
	}); err != nil {
		t.Fatal(err)
	}
	info, err := bot.GetWebhookInfo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if info.URL != webhookURL {
		t.Fatalf("webhook URL = %q, want %q", info.URL, webhookURL)
	}
	if keepWebhook {
		return
	}
	if err := bot.DeleteWebhook(ctx, hermes.DeleteWebhookParams{}); err != nil {
		t.Fatal(err)
	}
}
