package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendInvoiceContract(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/botTOKEN/sendInvoice" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		var params SendInvoiceParams
		if err := json.NewDecoder(request.Body).Decode(&params); err != nil {
			t.Fatal(err)
		}
		if params.Currency != "XTR" || len(params.Prices) != 1 || params.Prices[0].Amount != 25 {
			t.Fatalf("params = %#v", params)
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"message_id":1,"date":1,"chat":{"id":7,"type":"private"},"invoice":{"title":"Pro","description":"Upgrade","start_parameter":"","currency":"XTR","total_amount":25}}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	message, err := client.SendInvoice(context.Background(), SendInvoiceParams{
		ChatID: 7, Title: "Pro", Description: "Upgrade", Payload: "order-1",
		Currency: "XTR", Prices: []LabeledPrice{{Label: "Pro", Amount: 25}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if message.Invoice == nil || message.Invoice.TotalAmount != 25 {
		t.Fatalf("message = %#v", message)
	}
}

func TestInvoiceValidation(t *testing.T) {
	t.Parallel()

	client := New("TOKEN")
	_, err := client.SendInvoice(context.Background(), SendInvoiceParams{
		ChatID: 1, Title: "x", Description: "x", Payload: "x", Currency: "USD",
		Prices:       []LabeledPrice{{Label: "x", Amount: 1}},
		MaxTipAmount: 10, SuggestedTipAmounts: []int{5, 4},
	})
	if err == nil {
		t.Fatal("expected invalid tip order")
	}
	_, err = client.CreateInvoiceLink(context.Background(), CreateInvoiceLinkParams{
		Title: "x", Description: "x", Payload: "x", Currency: "USD",
		Prices: []LabeledPrice{{Label: "x", Amount: 1}}, SubscriptionPeriod: 2_592_000,
	})
	if err == nil {
		t.Fatal("expected invalid subscription currency")
	}
}

func TestSendPaidMediaStreamsAllAttachments(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		media := request.FormValue("media")
		if !strings.Contains(media, `"type":"live_photo"`) || !strings.Contains(media, `attach://clip`) || !strings.Contains(media, `attach://still`) {
			t.Fatalf("media = %s", media)
		}
		for _, field := range []string{"clip", "still"} {
			file, _, err := request.FormFile(field)
			if err != nil {
				t.Fatal(err)
			}
			_ = file.Close()
		}
		_, _ = io.WriteString(writer, `{"ok":true,"result":{"message_id":1,"date":1,"chat":{"id":7,"type":"private"}}}`)
	}))
	defer server.Close()

	client := New("TOKEN", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	_, err := client.SendPaidMediaUpload(context.Background(), SendPaidMediaParams{
		ChatID: 7, StarCount: 10,
		Media: []InputPaidMedia{InputPaidMediaLivePhoto{
			Media: Attachment("clip"), Photo: Attachment("still"),
		}},
	},
		NewUpload("clip", "clip.mp4", strings.NewReader("video")),
		NewUpload("still", "still.jpg", strings.NewReader("photo")),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPaymentQueryValidation(t *testing.T) {
	t.Parallel()

	client := New("TOKEN")
	if err := client.AnswerShippingQuery(context.Background(), AnswerShippingQueryParams{ShippingQueryID: "q", OK: true}); err == nil {
		t.Fatal("expected missing shipping options")
	}
	if err := client.AnswerPreCheckoutQuery(context.Background(), AnswerPreCheckoutQueryParams{PreCheckoutQueryID: "q"}); err == nil {
		t.Fatal("expected missing error message")
	}
}
