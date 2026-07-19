package benchmarks_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/sidwiskers/hermes"
	"github.com/sidwiskers/hermes/api"
	runtimecore "github.com/sidwiskers/hermes/internal/runtime"
	telegram "github.com/sidwiskers/hermes/types"
)

var updateJSON = func() []byte {
	data, err := os.ReadFile("testdata/update.json")
	if err != nil {
		panic(err)
	}
	return data
}()
var updateBatchJSON = func() []byte {
	items := make([]string, 100)
	for index := range items {
		items[index] = string(updateJSON)
	}
	return []byte("[" + strings.Join(items, ",") + "]")
}()

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

func BenchmarkDecodeUpdate(b *testing.B) {
	b.Run("fast", func(b *testing.B) {
		b.ReportAllocs()
		for index := 0; index < b.N; index++ {
			update, err := telegram.DecodeUpdate(updateJSON, false)
			if err != nil || update.UpdateID == 0 {
				b.Fatal(err)
			}
		}
	})
	b.Run("preserve_raw", func(b *testing.B) {
		b.ReportAllocs()
		for index := 0; index < b.N; index++ {
			update, err := telegram.DecodeUpdate(updateJSON, true)
			if err != nil || len(update.Raw) == 0 {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCompetitorDecodeUpdate(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		update, err := telegram.DecodeUpdate(updateJSON, false)
		if err != nil || update.UpdateID == 0 {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeUpdateBatch100(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		updates, err := telegram.DecodeUpdates(updateBatchJSON, false)
		if err != nil || len(updates) != 100 {
			b.Fatal(err)
		}
	}
}

func BenchmarkWebhookDecode(b *testing.B) {
	for _, test := range []struct {
		name     string
		preserve bool
	}{
		{name: "fast"},
		{name: "preserve_raw", preserve: true},
	} {
		b.Run(test.name, func(b *testing.B) {
			handler := runtimecore.WebhookHandler(
				runtimecore.WebhookOptions{PreserveRawUpdate: test.preserve},
				func(context.Context, *telegram.Update, bool) bool { return true },
			)
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				request := httptest.NewRequest(http.MethodPost, "/telegram", bytes.NewReader(updateJSON))
				response := httptest.NewRecorder()
				handler.ServeHTTP(response, request)
				if response.Code != http.StatusOK {
					b.Fatalf("status = %d", response.Code)
				}
			}
		})
	}
}

func routerBenchmarkFixture(routeCount, middlewareDepth int, pooling bool) (*hermes.Bot, *hermes.Update) {
	options := []hermes.Option{hermes.WithBotUsername("bench_bot"), hermes.WithContextPooling(pooling)}
	bot := hermes.New("TOKEN", options...)
	for index := 0; index < middlewareDepth; index++ {
		bot.Use(func(next hermes.Handler) hermes.Handler {
			return func(c *hermes.Context) error { return next(c) }
		})
	}
	for index := 0; index < routeCount-1; index++ {
		bot.Command("unused"+strconv.Itoa(index), func(*hermes.Context) error { return nil })
	}
	bot.Command("target", func(*hermes.Context) error { return nil })
	update := &hermes.Update{Message: &hermes.Message{
		MessageID: 42,
		Chat:      hermes.Chat{ID: 123, Type: "private"},
		Text:      "/target benchmark",
	}}
	return bot, update
}

func runRouterBenchmark(b *testing.B, bot *hermes.Bot, update *hermes.Update) {
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if err := bot.Handle(ctx, update); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRouterExact1(b *testing.B) {
	bot, update := routerBenchmarkFixture(1, 0, true)
	runRouterBenchmark(b, bot, update)
}

func BenchmarkRouterExact1PoolDisabled(b *testing.B) {
	bot, update := routerBenchmarkFixture(1, 0, false)
	runRouterBenchmark(b, bot, update)
}

func BenchmarkRouterExact1000(b *testing.B) {
	bot, update := routerBenchmarkFixture(1000, 0, true)
	runRouterBenchmark(b, bot, update)
}

func BenchmarkRouterMiddleware10(b *testing.B) {
	bot, update := routerBenchmarkFixture(1, 10, true)
	runRouterBenchmark(b, bot, update)
}

func BenchmarkRouterStartup1000(b *testing.B) {
	ctx := context.Background()
	update := &hermes.Update{Message: &hermes.Message{
		Chat: hermes.Chat{ID: 123, Type: "private"},
		Text: "/target",
	}}
	b.ReportAllocs()
	for run := 0; run < b.N; run++ {
		bot := hermes.New("TOKEN")
		for index := 0; index < 999; index++ {
			bot.Command("unused"+strconv.Itoa(index), func(*hermes.Context) error { return nil })
		}
		bot.Command("target", func(*hermes.Context) error { return nil })
		if err := bot.Handle(ctx, update); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAPICall(b *testing.B) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		_, _ = io.Copy(io.Discard, request.Body)
		_ = request.Body.Close()
		return response(`{"ok":true,"result":{"message_id":1,"date":1,"chat":{"id":123,"type":"private"},"text":"hello"}}`), nil
	})
	client := api.New("TOKEN", api.WithHTTPClient(&http.Client{Transport: transport}))
	params := api.SendMessageParams{ChatID: int64(123), Text: "hello"}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		message, err := client.SendMessage(ctx, params)
		if err != nil || message.MessageID != 1 {
			b.Fatal(err)
		}
	}
}

func BenchmarkMultipartUpload1MiB(b *testing.B) {
	payload := bytes.Repeat([]byte("x"), 1<<20)
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		_, _ = io.Copy(io.Discard, request.Body)
		_ = request.Body.Close()
		return response(`{"ok":true,"result":{"message_id":1,"date":1,"chat":{"id":123,"type":"private"},"photo":[]}}`), nil
	})
	client := api.New("TOKEN", api.WithHTTPClient(&http.Client{Transport: transport}))
	ctx := context.Background()

	b.SetBytes(int64(len(payload)))
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, err := client.SendPhotoUpload(ctx, api.SendPhotoParams{ChatID: int64(123)}, "image.jpg", bytes.NewReader(payload))
		if err != nil {
			b.Fatal(err)
		}
	}
}
