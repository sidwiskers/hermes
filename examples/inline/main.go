package main

import (
	"context"
	"log"
	"os"

	"github.com/sidwiskers/hermes"
)

func main() {
	bot := hermes.New(os.Getenv("BOT_TOKEN"))
	bot.Use(hermes.Recover())

	bot.On(hermes.UpdateIs(hermes.UpdateInlineQuery), func(c *hermes.Context) error {
		query := c.Update.InlineQuery
		return c.Bot.AnswerInlineQuery(c, hermes.AnswerInlineQueryParams{
			InlineQueryID: query.ID,
			CacheTime:     10,
			Results: []hermes.InlineQueryResult{
				hermes.InlineQueryResultArticle{
					ID:    "hello",
					Title: "Say hello",
					InputMessageContent: hermes.InputTextMessageContent{
						MessageText: "Hello from Hermes.",
					},
				},
			},
		})
	})

	if err := bot.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
