package gotelegrambot_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	telegrambot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const messageResponse = `{"ok":true,"result":{"message_id":1,"date":1,"chat":{"id":123,"type":"private"},"text":"hello"}}`

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) Do(request *http.Request) (*http.Response, error) {
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
		var update models.Update
		if err := json.Unmarshal(updateJSON, &update); err != nil {
			b.Fatal(err)
		}
	}
}

func newRouter(b *testing.B, routes, middleware int) (*telegrambot.Bot, *models.Update) {
	b.Helper()
	middlewares := make([]telegrambot.Middleware, middleware)
	for index := range middlewares {
		middlewares[index] = func(next telegrambot.HandlerFunc) telegrambot.HandlerFunc {
			return func(ctx context.Context, bot *telegrambot.Bot, update *models.Update) {
				next(ctx, bot, update)
			}
		}
	}
	bot, err := telegrambot.New("1:TOKEN", telegrambot.WithSkipGetMe(), telegrambot.WithNotAsyncHandlers(), telegrambot.WithMiddlewares(middlewares...))
	if err != nil {
		b.Fatal(err)
	}
	for index := range routes - 1 {
		bot.RegisterHandler(telegrambot.HandlerTypeMessageText, "unused"+strconv.Itoa(index), telegrambot.MatchTypeCommandStartOnly,
			func(context.Context, *telegrambot.Bot, *models.Update) {})
	}
	bot.RegisterHandler(telegrambot.HandlerTypeMessageText, "start", telegrambot.MatchTypeCommandStartOnly,
		func(context.Context, *telegrambot.Bot, *models.Update) {})
	var update models.Update
	if err := json.Unmarshal(updateJSON, &update); err != nil {
		b.Fatal(err)
	}
	return bot, &update
}

func runRouter(b *testing.B, bot *telegrambot.Bot, update *models.Update) {
	b.Helper()
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		bot.ProcessUpdate(ctx, update)
	}
}

func BenchmarkRouterExact1(b *testing.B) {
	bot, update := newRouter(b, 1, 0)
	runRouter(b, bot, update)
}

func BenchmarkRouterExact1000(b *testing.B) {
	bot, update := newRouter(b, 1000, 0)
	runRouter(b, bot, update)
}

func BenchmarkRouterMiddleware10(b *testing.B) {
	bot, update := newRouter(b, 1, 10)
	runRouter(b, bot, update)
}

func BenchmarkAPICall(b *testing.B) {
	client := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		_, _ = io.Copy(io.Discard, request.Body)
		_ = request.Body.Close()
		return response(messageResponse), nil
	})
	bot, err := telegrambot.New(
		"1:TOKEN",
		telegrambot.WithSkipGetMe(),
		telegrambot.WithHTTPClient(time.Minute, client),
	)
	if err != nil {
		b.Fatal(err)
	}
	params := &telegrambot.SendMessageParams{ChatID: int64(123), Text: "hello"}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		message, err := bot.SendMessage(ctx, params)
		if err != nil || message.ID != 1 {
			b.Fatal(err)
		}
	}
}
