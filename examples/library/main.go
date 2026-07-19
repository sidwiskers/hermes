package main

import (
	"context"
	"log"
	"os"

	"github.com/sidwiskers/hermes/api"
)

func main() {
	client := api.New(os.Getenv("BOT_TOKEN"))
	_, err := client.SendMessage(context.Background(), api.SendMessageParams{
		ChatID: os.Getenv("CHAT_ID"),
		Text:   "Hello from the standalone API client.",
	})
	if err != nil {
		log.Fatal(err)
	}
}
