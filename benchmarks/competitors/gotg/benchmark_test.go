package gotg_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	tg "github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
)

const (
	messageResponse = `{"ok":true,"result":{"message_id":1,"date":1,"chat":{"id":123,"type":"private"},"text":"hello"}}`
	meResponse      = `{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Bench","username":"bench_bot"}}`
)

type doerFunc func(*http.Request) (*http.Response, error)

func (fn doerFunc) Do(request *http.Request) (*http.Response, error) { return fn(request) }

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
		var update tg.Update
		if err := json.Unmarshal(updateJSON, &update); err != nil {
			b.Fatal(err)
		}
	}
}

func newRouter(b *testing.B, routes, middleware int) (*tgb.Router, *tgb.Update) {
	b.Helper()
	client := tg.New("1:TOKEN", tg.WithClientDoer(doerFunc(func(request *http.Request) (*http.Response, error) {
		_, _ = io.Copy(io.Discard, request.Body)
		_ = request.Body.Close()
		return response(meResponse), nil
	})))
	if _, err := client.Me(context.Background()); err != nil {
		b.Fatal(err)
	}
	router := tgb.NewRouter()
	for range middleware {
		router.Use(tgb.MiddlewareFunc(func(next tgb.Handler) tgb.Handler {
			return tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
				return next.Handle(ctx, update)
			})
		}))
	}
	for index := range routes - 1 {
		router.Message(func(context.Context, *tgb.MessageUpdate) error { return nil },
			tgb.Command("unused"+strconv.Itoa(index)))
	}
	router.Message(func(context.Context, *tgb.MessageUpdate) error { return nil }, tgb.Command("start"))
	var update tg.Update
	if err := json.Unmarshal(updateJSON, &update); err != nil {
		b.Fatal(err)
	}
	return router, &tgb.Update{Update: &update, Client: client}
}

func runRouter(b *testing.B, router *tgb.Router, update *tgb.Update) {
	b.Helper()
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := router.Handle(ctx, update); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRouterExact1(b *testing.B) {
	router, update := newRouter(b, 1, 0)
	runRouter(b, router, update)
}

func BenchmarkRouterExact1000(b *testing.B) {
	router, update := newRouter(b, 1000, 0)
	runRouter(b, router, update)
}

func BenchmarkRouterMiddleware10(b *testing.B) {
	router, update := newRouter(b, 1, 10)
	runRouter(b, router, update)
}

func BenchmarkAPICall(b *testing.B) {
	client := tg.New("1:TOKEN", tg.WithClientDoer(doerFunc(func(request *http.Request) (*http.Response, error) {
		_, _ = io.Copy(io.Discard, request.Body)
		_ = request.Body.Close()
		return response(messageResponse), nil
	})))
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		message, err := client.SendMessage(tg.ChatID(123), "hello").Do(ctx)
		if err != nil || message.ID != 1 {
			b.Fatal(err)
		}
	}
}
