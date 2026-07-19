package tgbotapi_test

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
		var update tgbotapi.Update
		if err := json.Unmarshal(updateJSON, &update); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAPICall(b *testing.B) {
	client := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		_, _ = io.Copy(io.Discard, request.Body)
		_ = request.Body.Close()
		if strings.HasSuffix(request.URL.Path, "/getMe") {
			return response(`{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"bench","username":"bench_bot"}}`), nil
		}
		return response(messageResponse), nil
	})
	bot, err := tgbotapi.NewBotAPIWithClient("1:TOKEN", tgbotapi.APIEndpoint, client)
	if err != nil {
		b.Fatal(err)
	}
	params := tgbotapi.NewMessage(123, "hello")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		message, err := bot.Send(params)
		if err != nil || message.MessageID != 1 {
			b.Fatal(err)
		}
	}
}
