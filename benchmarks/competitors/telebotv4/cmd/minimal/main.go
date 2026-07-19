package main

import tele "gopkg.in/telebot.v4"

func main() {
	_, _ = tele.NewBot(tele.Settings{Offline: true})
}
