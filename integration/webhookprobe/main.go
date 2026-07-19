//go:build integration

// Command webhookprobe runs the disposable server used by the live webhook
// release gate. It must never be deployed as an application service.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sidwiskers/hermes"
	"github.com/sidwiskers/hermes/dedupe"
)

var errIntentionalRetry = errors.New("intentional webhook retry probe")

func main() {
	token := os.Getenv("HERMES_TEST_BOT_TOKEN")
	secret := os.Getenv("HERMES_TEST_WEBHOOK_SECRET")
	if token == "" || secret == "" {
		log.Fatal("HERMES_TEST_BOT_TOKEN and HERMES_TEST_WEBHOOK_SECRET are required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	bot := hermes.New(token, hermes.WithMaxConcurrentUpdates(32))
	claims := dedupe.NewMemory(dedupe.MemoryConfig{MaxEntries: 4096, Shards: 16})
	bot.Use(dedupe.New(claims, dedupe.WithNamespace("webhook-probe")).Middleware())
	bot.Use(hermes.Recover())

	bot.Command("start", directReply("Hermes webhook delivery and direct reply passed."))

	var retryAttempts sync.Map
	bot.Command("retry", func(c *hermes.Context) error {
		if _, retried := retryAttempts.LoadOrStore(c.Update.UpdateID, struct{}{}); !retried {
			log.Print("RETRY_PROBE_FIRST_ATTEMPT")
			return errIntentionalRetry
		}
		log.Print("RETRY_PROBE_RECOVERED")
		return respond(c, "Hermes webhook retry passed.")
	})

	var panicAttempts sync.Map
	bot.Command("panic", func(c *hermes.Context) error {
		if _, retried := panicAttempts.LoadOrStore(c.Update.UpdateID, struct{}{}); !retried {
			log.Print("PANIC_PROBE_FIRST_ATTEMPT")
			panic("intentional webhook panic probe")
		}
		log.Print("PANIC_PROBE_RECOVERED")
		return respond(c, "Hermes webhook panic recovery passed.")
	})

	log.Print("WEBHOOK_PROBE_READY address=:8080 path=/telegram")
	if err := bot.ServeWebhookReplies(ctx, ":8080", "/telegram", hermes.WebhookOptions{
		Secret:           secret,
		MaxResponseBytes: 64 << 10,
	}); err != nil {
		log.Fatal(err)
	}
	log.Print("WEBHOOK_PROBE_DRAINED")
}

func directReply(text string) hermes.Handler {
	return func(c *hermes.Context) error {
		return respond(c, text)
	}
}

func respond(c *hermes.Context, text string) error {
	chatID, ok := c.ChatID()
	if !ok {
		return errors.New("webhook probe update has no chat")
	}
	return c.RespondWebhook("sendMessage", hermes.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
}
