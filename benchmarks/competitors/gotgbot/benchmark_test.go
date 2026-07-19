package gotgbot_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

const messageResponse = `{"ok":true,"result":{"message_id":1,"date":1,"chat":{"id":123,"type":"private"},"text":"hello"}}`

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func response(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

var updateJSON = mustReadFixture()

func mustReadFixture() []byte {
	data, err := os.ReadFile("../../testdata/update.json")
	if err != nil {
		panic(err)
	}
	return data
}

func BenchmarkCompetitorDecodeUpdate(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var update gotgbot.Update
		if err := json.Unmarshal(updateJSON, &update); err != nil {
			b.Fatal(err)
		}
	}
}

func newRouter(b *testing.B, routes int) (*ext.Dispatcher, *gotgbot.Bot, *gotgbot.Update) {
	b.Helper()
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{MaxRoutines: -1})
	for index := range routes - 1 {
		dispatcher.AddHandler(handlers.NewCommand("unused"+strconv.Itoa(index), func(*gotgbot.Bot, *ext.Context) error { return nil }))
	}
	dispatcher.AddHandler(handlers.NewCommand("start", func(*gotgbot.Bot, *ext.Context) error { return nil }))
	bot := &gotgbot.Bot{User: gotgbot.User{Username: "bench_bot"}}
	var update gotgbot.Update
	if err := json.Unmarshal(updateJSON, &update); err != nil {
		b.Fatal(err)
	}
	return dispatcher, bot, &update
}

func runRouter(b *testing.B, dispatcher *ext.Dispatcher, bot *gotgbot.Bot, update *gotgbot.Update) {
	b.Helper()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := dispatcher.ProcessUpdate(bot, update, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRouterExact1(b *testing.B) {
	dispatcher, bot, update := newRouter(b, 1)
	runRouter(b, dispatcher, bot, update)
}

func BenchmarkRouterExact1000(b *testing.B) {
	dispatcher, bot, update := newRouter(b, 1000)
	runRouter(b, dispatcher, bot, update)
}

func BenchmarkAPICall(b *testing.B) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		_, _ = io.Copy(io.Discard, request.Body)
		_ = request.Body.Close()
		return response(messageResponse), nil
	})
	client := &gotgbot.BaseBotClient{Client: http.Client{Transport: transport}}
	bot, err := gotgbot.NewBot("1:TOKEN", &gotgbot.BotOpts{
		BotClient:         client,
		DisableTokenCheck: true,
	})
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		message, err := bot.SendMessageWithContext(ctx, 123, "hello", nil)
		if err != nil || message.MessageId != 1 {
			b.Fatal(err)
		}
	}
}
