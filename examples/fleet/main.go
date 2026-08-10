package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/sidwiskers/hermes"
	"github.com/sidwiskers/hermes/fleet"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	host := fleet.New(fleet.WithBotMaxConcurrentUpdates(32))

	alerts := host.NewBot(os.Getenv("ALERTS_BOT_TOKEN"))
	alerts.Command("start", func(c *hermes.Context) error {
		return c.Send("Alerts bot is ready.")
	})

	support := host.NewBot(os.Getenv("SUPPORT_BOT_TOKEN"))
	support.Command("start", func(c *hermes.Context) error {
		return c.Send("Support bot is ready.")
	})

	if err := host.Mount("alerts", alerts); err != nil {
		log.Fatal(err)
	}
	if err := host.Mount("support", support); err != nil {
		log.Fatal(err)
	}
	if err := host.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
