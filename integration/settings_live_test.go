//go:build integration

package integration

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sidwiskers/hermes"
)

// TestLiveSetMySettings exercises Telegram's production setMySettings
// endpoint using a dedicated disposable bot. It cycles through every
// callback_query_mode value and leaves the bot in the legacy-compatible
// "default" mode. No messages, chats, or webhook state are touched.
func TestLiveSetMySettings(t *testing.T) {
	token := strings.TrimSpace(os.Getenv("HERMES_TEST_BOT_TOKEN"))
	if token == "" {
		t.Skip("HERMES_TEST_BOT_TOKEN is required")
	}

	bot := hermes.New(token)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	me, err := bot.GetMe(ctx)
	if err != nil {
		t.Fatalf("getMe failed before setMySettings: %v", err)
	}
	if !me.IsBot || me.ID == 0 || me.Username == "" {
		t.Fatalf("token did not resolve to a valid bot identity: %#v", me)
	}
	if expected := strings.TrimPrefix(strings.TrimSpace(os.Getenv("HERMES_TEST_BOT_USERNAME")), "@"); expected != "" && !strings.EqualFold(me.Username, expected) {
		t.Fatalf("bot username = %q, want %q", me.Username, expected)
	}
	t.Logf("authenticated dedicated test bot @%s", me.Username)

	restored := false
	t.Cleanup(func() {
		if restored {
			return
		}
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()
		defaultMode := "default"
		if cleanupErr := bot.SetMySettings(cleanupCtx, hermes.SetMySettingsParams{
			CallbackQueryMode: &defaultMode,
		}); cleanupErr != nil {
			t.Errorf("failed to restore callback_query_mode to default: %v", cleanupErr)
		}
	})

	for _, mode := range []string{"auto", "inline", "default"} {
		selectedMode := mode
		if err := bot.SetMySettings(ctx, hermes.SetMySettingsParams{
			CallbackQueryMode: &selectedMode,
		}); err != nil {
			t.Fatalf("setMySettings callback_query_mode=%q failed: %v", mode, err)
		}
		t.Logf("Telegram accepted callback_query_mode=%q", mode)
	}
	restored = true
}
