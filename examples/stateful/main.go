package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sidwiskers/hermes"
	"github.com/sidwiskers/hermes/fsm"
	"github.com/sidwiskers/hermes/session"
)

type state uint8

const (
	idle state = iota
	waitingForName
)

type profile struct {
	Name string
}

func main() {
	bot := hermes.New(os.Getenv("BOT_TOKEN"))

	store := session.NewMemory[fsm.Snapshot[state, profile]](24 * time.Hour)
	sessions := session.New(store, session.ByChatUser, session.WithNamespace("profile"))
	flow := fsm.New(sessions, idle)
	if err := flow.Add(fsm.Rule[state, profile]{From: idle, Event: "begin", To: waitingForName}); err != nil {
		log.Fatal(err)
	}

	bot.Use(hermes.Recover(), flow.Middleware())
	bot.Command("profile", flow.Then("begin", func(c *hermes.Context) error {
		return c.Send("What should I call you?")
	}))
	bot.On(flow.In(waitingForName), func(c *hermes.Context) error {
		name := strings.TrimSpace(c.Text())
		if name == "" {
			return c.Send("Please send a name.")
		}
		if err := flow.Set(c, fsm.Snapshot[state, profile]{State: idle, Data: profile{Name: name}}); err != nil {
			return err
		}
		return c.Send("Saved, " + name + ".")
	})

	if err := bot.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
