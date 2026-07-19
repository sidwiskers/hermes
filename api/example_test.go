package api_test

import (
	"context"
	"fmt"

	"github.com/sidwiskers/hermes/api"
	"github.com/sidwiskers/hermes/testkit"
)

func ExampleClient_SendMessage() {
	client, recorder := testkit.NewClient()
	recorder.Respond(api.Message{MessageID: 9, Chat: api.Chat{ID: 7, Type: "private"}})

	message, err := client.SendMessage(context.Background(), api.SendMessageParams{
		ChatID: 7,
		Text:   "hello",
	})
	if err != nil {
		panic(err)
	}

	request, _ := recorder.Last()
	fmt.Println(message.MessageID, request.Method)
	// Output: 9 sendMessage
}
