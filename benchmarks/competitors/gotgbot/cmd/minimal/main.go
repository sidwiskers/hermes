package main

import "github.com/PaulSonOfLars/gotgbot/v2"

func main() {
	_, _ = gotgbot.NewBot("1:TOKEN", &gotgbot.BotOpts{DisableTokenCheck: true})
}
