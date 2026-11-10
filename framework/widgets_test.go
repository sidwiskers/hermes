package framework

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sidwiskers/hermes/api"
	telegram "github.com/sidwiskers/hermes/types"
)

func TestWidgetQueryContextAndAnswer(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["widget_query_id"] != "widget-query-1" || body["data"] != "fresh-state" {
			t.Fatalf("answerWidgetQuery payload = %#v", body)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":true}`)
	}))
	defer server.Close()

	client := api.New("TOKEN", api.WithBaseURL(server.URL), api.WithHTTPClient(server.Client()))
	update := &telegram.Update{
		WidgetQuery: &telegram.WidgetQuery{
			ID:   "widget-query-1",
			From: telegram.User{ID: 42, FirstName: "Widget"},
		},
	}
	ctx := NewContext(context.Background(), client, update, "")
	if ctx.Type() != UpdateWidgetQuery {
		t.Fatalf("context type = %q", ctx.Type())
	}
	if !UpdateIs(UpdateWidgetQuery)(ctx) {
		t.Fatal("UpdateIs did not match widget query")
	}
	if ctx.WidgetQuery == nil || ctx.Sender() == nil || ctx.Sender().ID != 42 {
		t.Fatalf("widget query context = %#v, sender = %#v", ctx.WidgetQuery, ctx.Sender())
	}
	if err := ctx.AnswerWidgetQuery("fresh-state"); err != nil {
		t.Fatal(err)
	}
}

func TestSharedWidgetQueryContext(t *testing.T) {
	t.Parallel()

	update := &telegram.Update{
		SharedWidgetQuery: &telegram.SharedWidgetQuery{
			ID:   "shared-query-1",
			From: telegram.User{ID: 99, FirstName: "Shared"},
		},
	}
	ctx := NewContext(context.Background(), nil, update, "")
	if ctx.Type() != UpdateSharedWidgetQuery || !UpdateIs(UpdateSharedWidgetQuery)(ctx) {
		t.Fatalf("shared widget query update type = %q", ctx.Type())
	}
	if ctx.SharedWidgetQuery == nil || ctx.Sender() == nil || ctx.Sender().ID != 99 {
		t.Fatalf("shared widget query context = %#v, sender = %#v", ctx.SharedWidgetQuery, ctx.Sender())
	}
	if err := ctx.AnswerWidgetQuery("ignored"); err == nil {
		t.Fatal("expected missing-client validation error")
	}
}
