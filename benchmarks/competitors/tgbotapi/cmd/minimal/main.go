package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func main() {
	_, _ = tgbotapi.NewBotAPI("1:TOKEN")
}
