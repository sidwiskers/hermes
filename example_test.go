package hermes_test

import (
	"context"
	"fmt"

	"github.com/sidwiskers/hermes"
	"github.com/sidwiskers/hermes/testkit"
)

func Example() {
	bot, recorder := testkit.New()
	recorder.Respond(hermes.Message{MessageID: 2, Chat: hermes.Chat{ID: 1, Type: "private"}})

	bot.Command("start", func(c *hermes.Context) error {
		return c.Send("Hello from Hermes")
	})

	err := bot.Handle(context.Background(), &hermes.Update{
		UpdateID: 1,
		Message: &hermes.Message{
			MessageID: 1,
			Chat:      hermes.Chat{ID: 1, Type: "private"},
			Text:      "/start",
		},
	})
	if err != nil {
		panic(err)
	}

	request, _ := recorder.Last()
	fmt.Println(request.Method, request.JSON["text"])
	// Output: sendMessage Hello from Hermes
}

func ExampleCallbackCodec() {
	users := hermes.Int64Callback("user:")
	data, err := users.Data(42)
	if err != nil {
		panic(err)
	}

	userID, err := users.Parse(data)
	if err != nil {
		panic(err)
	}
	fmt.Println(data, userID)
	// Output: user:42 42
}
