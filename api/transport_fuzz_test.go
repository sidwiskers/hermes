package api

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

type fuzzRoundTripper func(*http.Request) (*http.Response, error)

func (fn fuzzRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func FuzzTelegramEnvelopeDecode(f *testing.F) {
	f.Add([]byte(`{"ok":true,"result":{"value":1}}`), 200)
	f.Add([]byte(`{"ok":false,"error_code":429,"description":"retry","parameters":{"retry_after":1}}`), 200)
	f.Add([]byte(`not-json`), 502)

	f.Fuzz(func(t *testing.T, body []byte, status int) {
		if status < 100 || status > 599 {
			status = 200
		}
		const token = "SECRET_TOKEN_MUST_NOT_LEAK"
		client := New(token, WithHTTPClient(&http.Client{Transport: fuzzRoundTripper(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: status,
				Status:     http.StatusText(status),
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(string(body))),
			}, nil
		})}))
		var result map[string]any
		err := client.Call(context.Background(), "testMethod", nil, &result)
		if err != nil && strings.Contains(err.Error(), token) {
			t.Fatal("transport error leaked bot token")
		}
	})
}

func FuzzMethodValidation(f *testing.F) {
	f.Add("sendMessage")
	f.Add("edit_message")
	f.Add("../../token")
	f.Fuzz(func(t *testing.T, method string) {
		valid := validMethod(method)
		if valid {
			for _, char := range method {
				allowed := char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '_'
				if !allowed {
					t.Fatalf("invalid method accepted: %q", method)
				}
			}
		}
	})
}

func FuzzMultipartHeaderValidation(f *testing.F) {
	f.Add("photo", "image.jpg")
	f.Add("photo\r\nX-Injected", "image.jpg")
	f.Add("photo", "image.jpg\nX-Injected")
	f.Fuzz(func(t *testing.T, field, name string) {
		err := validateMultipartInputs(nil, []Upload{{
			Field: field, Name: name, Reader: strings.NewReader("x"),
		}})
		valid := validMultipartHeaderValue(field) && (name == "" || validMultipartHeaderValue(filepath.Base(name)))
		if valid == (err != nil) {
			t.Fatalf("field=%q name=%q valid=%v error=%v", field, name, valid, err)
		}
	})
}
