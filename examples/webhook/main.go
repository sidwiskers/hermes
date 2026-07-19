package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/sidwiskers/hermes"
)

func main() {
	token := required("BOT_TOKEN")
	publicURL := strings.TrimRight(required("WEBHOOK_URL"), "/")
	secret := required("WEBHOOK_SECRET")
	address := os.Getenv("LISTEN_ADDRESS")
	if address == "" {
		address = ":8080"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	bot := hermes.New(token)
	bot.Use(hermes.Recover())
	bot.Command("start", func(c *hermes.Context) error {
		return c.Send("Webhook is ready.")
	})

	const path = "/telegram"
	if err := bot.SetWebhook(ctx, hermes.SetWebhookParams{
		URL:         publicURL + path,
		SecretToken: secret,
	}); err != nil {
		log.Fatal(err)
	}

	if err := bot.ServeWebhook(ctx, address, path, hermes.WebhookOptions{
		Secret: secret,
	}); err != nil {
		log.Fatal(err)
	}
}

func required(name string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		log.Fatalf("%s is required", name)
	}
	return value
}
