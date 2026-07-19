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

	user := hermes.Int64Callback("user:")
	private := bot.Group(hermes.PrivateChat)

	private.Command("start", func(c *hermes.Context) error {
		keyboard := hermes.Keyboard(hermes.Row(user.MustButton("Open user", 42)))
		return c.Send("Ready.", hermes.WithKeyboard(keyboard))
	})

	bot.CallbackPrefix(user.Prefix, user.Handler(func(c *hermes.Context, userID int64) error {
		if err := c.Acknowledge(); err != nil {
			return err
		}
		return c.Edit(fmt.Sprintf("Selected user %d", userID))
	}))

	bot.On(hermes.All(hermes.GroupChat, hermes.PhotoMessage), func(c *hermes.Context) error {
		return c.Ephemeral("Photo received privately.")
	})

	bot.Command("poll", func(c *hermes.Context) error {
		return c.Poll("Choose one", "Alpha", "Beta")
	})

	bot.Command("dice", func(c *hermes.Context) error {
		return c.Dice(hermes.DiceEmoji)
	})

	bot.On(hermes.UpdateIs(hermes.UpdateChatJoinRequest), func(c *hermes.Context) error {
		return c.ApproveJoinRequest()
	})

	if err := bot.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
