package main

import telegrambot "github.com/go-telegram/bot"

func main() {
	_, _ = telegrambot.New("1:TOKEN", telegrambot.WithSkipGetMe())
}
