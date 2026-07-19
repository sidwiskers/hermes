package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/sidwiskers/hermes"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	bot := hermes.New(os.Getenv("BOT_TOKEN"))
	bot.Use(hermes.Recover())
	bot.Command("start", func(c *hermes.Context) error {
		chatID, ok := c.ChatID()
		if !ok {
			return nil
		}
		return c.RespondWebhook("sendMessage", hermes.SendMessageParams{
			ChatID: chatID,
			Text:   "Acknowledged in the webhook response.",
		})
	})

	if err := bot.ServeWebhookReplies(ctx, ":8080", "/telegram", hermes.WebhookOptions{
		Secret: os.Getenv("WEBHOOK_SECRET"),
	}); err != nil {
		log.Fatal(err)
	}
}
