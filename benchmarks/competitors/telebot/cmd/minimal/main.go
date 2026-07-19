package main

import tele "gopkg.in/telebot.v3"

func main() {
	_, _ = tele.NewBot(tele.Settings{Offline: true})
}
