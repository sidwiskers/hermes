package telego_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	gojson "github.com/grbit/go-json"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

const (
	token           = "1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	messageResponse = `{"ok":true,"result":{"message_id":1,"date":1,"chat":{"id":123,"type":"private"},"text":"hello"}}`
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return fn(request) }

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
		var update telego.Update
		if err := gojson.Unmarshal(updateJSON, &update); err != nil {
			b.Fatal(err)
		}
	}
}

func newRouter(b *testing.B, routes, middleware int) (*telego.Bot, *telegohandler.HandlerGroup, telego.Update) {
	b.Helper()
	bot, err := telego.NewBot(token, telego.WithDiscardLogger())
	if err != nil {
		b.Fatal(err)
	}
	handler, err := telegohandler.NewBotHandler(bot, nil)
	if err != nil {
		b.Fatal(err)
	}
	group := handler.BaseGroup()
	for range middleware {
		group.Use(func(ctx *telegohandler.Context, update telego.Update) error {
			return ctx.Next(update)
		})
	}
	for index := range routes - 1 {
		group.Handle(func(*telegohandler.Context, telego.Update) error { return nil },
			telegohandler.CommandEqual("unused"+strconv.Itoa(index)))
	}
	group.Handle(func(*telegohandler.Context, telego.Update) error { return nil },
		telegohandler.CommandEqual("start"))
	var update telego.Update
	if err := gojson.Unmarshal(updateJSON, &update); err != nil {
		b.Fatal(err)
	}
	return bot, group, update
}

func runRouter(b *testing.B, bot *telego.Bot, group *telegohandler.HandlerGroup, update telego.Update) {
	b.Helper()
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := group.HandleUpdate(ctx, bot, update); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRouterExact1(b *testing.B) {
	bot, group, update := newRouter(b, 1, 0)
	runRouter(b, bot, group, update)
}

func BenchmarkRouterExact1000(b *testing.B) {
	bot, group, update := newRouter(b, 1000, 0)
	runRouter(b, bot, group, update)
}

func BenchmarkRouterMiddleware10(b *testing.B) {
	bot, group, update := newRouter(b, 1, 10)
	runRouter(b, bot, group, update)
}

func BenchmarkAPICall(b *testing.B) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		_, _ = io.Copy(io.Discard, request.Body)
		_ = request.Body.Close()
		return response(messageResponse), nil
	})
	bot, err := telego.NewBot(
		token,
		telego.WithDiscardLogger(),
		telego.WithHTTPClient(&http.Client{Transport: transport}),
	)
	if err != nil {
		b.Fatal(err)
	}
	params := &telego.SendMessageParams{ChatID: telego.ChatID{ID: 123}, Text: "hello"}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		message, err := bot.SendMessage(ctx, params)
		if err != nil || message.MessageID != 1 {
			b.Fatal(err)
		}
	}
}
