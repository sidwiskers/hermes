package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/sidwiskers/hermes"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	bot := hermes.New(
		os.Getenv("BOT_TOKEN"),
		hermes.WithMaxConcurrentUpdates(64),
		hermes.WithBotUsername(os.Getenv("BOT_USERNAME")),
	)

	bot.Use(
		hermes.Recover(),
		hermes.Timeout(15*time.Second),
	)

	bot.Command("start", func(c *hermes.Context) error {
		return c.Send("Hello from hermes.")
	})

	bot.Command("profile", func(c *hermes.Context) error {
		return c.Ephemeral(
			"<b>Private profile</b>\nOnly you can see this.",
			hermes.HTML,
			hermes.WithKeyboard(hermes.Keyboard(
				hermes.Row(hermes.Button("Refresh", "profile:refresh")),
			)),
		)
	})

	bot.CallbackPrefix("profile:", func(c *hermes.Context) error {
		return c.Answer("Updated")
	})

	if err := bot.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
