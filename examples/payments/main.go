package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sidwiskers/hermes"
)

func main() {
	bot := hermes.New(os.Getenv("BOT_TOKEN"))
	bot.Use(hermes.Recover())

	bot.Command("buy", func(c *hermes.Context) error {
		chatID, ok := c.ChatID()
		if !ok {
			return fmt.Errorf("buy command has no chat")
		}
		_, err := c.Bot.SendInvoice(c, hermes.SendInvoiceParams{
			ChatID:      chatID,
			Title:       "Hermes demo",
			Description: "A sample Telegram Stars purchase",
			Payload:     "hermes-demo-v1",
			Currency:    "XTR",
			Prices: []hermes.LabeledPrice{
				{Label: "Demo item", Amount: 1},
			},
		})
		return err
	})

	bot.On(hermes.UpdateIs(hermes.UpdatePreCheckoutQuery), func(c *hermes.Context) error {
		query := c.Update.PreCheckoutQuery
		return c.Bot.AnswerPreCheckoutQuery(c, hermes.AnswerPreCheckoutQueryParams{
			PreCheckoutQueryID: query.ID,
			OK:                 true,
		})
	})

	if err := bot.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
