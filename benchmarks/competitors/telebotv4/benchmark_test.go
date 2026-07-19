package telebotv4_test

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	tele "gopkg.in/telebot.v4"
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
		var update tele.Update
		if err := json.Unmarshal(updateJSON, &update); err != nil {
			b.Fatal(err)
		}
	}
}

func newRouter(b *testing.B, routes, middleware int) (*tele.Bot, tele.Update) {
	b.Helper()
	bot, err := tele.NewBot(tele.Settings{Offline: true, Synchronous: true})
	if err != nil {
		b.Fatal(err)
	}
	for range middleware {
		bot.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
			return func(context tele.Context) error { return next(context) }
		})
	}
	for index := range routes - 1 {
		bot.Handle("/unused"+strconv.Itoa(index), func(tele.Context) error { return nil })
	}
	bot.Handle("/start", func(tele.Context) error { return nil })
	var update tele.Update
	if err := json.Unmarshal(updateJSON, &update); err != nil {
		b.Fatal(err)
	}
	return bot, update
}

func runRouter(b *testing.B, bot *tele.Bot, update tele.Update) {
	b.Helper()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		bot.ProcessUpdate(update)
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
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		_, _ = io.Copy(io.Discard, request.Body)
		_ = request.Body.Close()
		return response(messageResponse), nil
	})
	bot, err := tele.NewBot(tele.Settings{
		Token:   "1:TOKEN",
		Offline: true,
		Client:  &http.Client{Transport: transport},
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		message, err := bot.Send(tele.ChatID(123), "hello")
		if err != nil || message.ID != 1 {
			b.Fatal(err)
		}
	}
}
