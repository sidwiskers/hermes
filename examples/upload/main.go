package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sidwiskers/hermes"
)

func main() {
	path := os.Getenv("DOCUMENT_PATH")
	if path == "" {
		log.Fatal("DOCUMENT_PATH is required")
	}

	bot := hermes.New(os.Getenv("BOT_TOKEN"))
	bot.Use(hermes.Recover())
	bot.Command("upload", func(c *hermes.Context) error {
		chatID, ok := c.ChatID()
		if !ok {
			return fmt.Errorf("upload command has no chat")
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = c.Bot.SendDocumentUpload(c, hermes.SendDocumentParams{
			ChatID:  chatID,
			Caption: "Streamed by Hermes",
		}, filepath.Base(path), file)
		return err
	})

	if err := bot.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
