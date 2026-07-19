package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNilClientMethodsReturnError(t *testing.T) {
	t.Parallel()

	var client *Client
	if err := client.Call(context.Background(), "getMe", nil, nil); !errors.Is(err, ErrClientRequired) {
		t.Fatalf("Call error = %v", err)
	}
	if err := client.CallMultipart(context.Background(), "sendPhoto", nil, nil, nil); !errors.Is(err, ErrClientRequired) {
		t.Fatalf("CallMultipart error = %v", err)
	}
	if _, err := Call[bool](context.Background(), client, "getMe", nil); !errors.Is(err, ErrClientRequired) {
		t.Fatalf("generic Call error = %v", err)
	}
	if _, err := client.OpenFile(context.Background(), "file"); !errors.Is(err, ErrClientRequired) {
		t.Fatalf("OpenFile error = %v", err)
	}
	if _, err := client.DownloadFile(context.Background(), "file", io.Discard); !errors.Is(err, ErrClientRequired) {
		t.Fatalf("DownloadFile error = %v", err)
	}
}

func TestClientTestEnvironmentMethodPath(t *testing.T) {
	t.Parallel()

	var path string
	client := New("TOKEN",
		WithBaseURL("https://telegram.invalid"),
		WithTestEnvironment(true),
		WithHTTPClient(&http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			path = request.URL.Path
			return testResponse(http.StatusOK, `{"ok":true,"result":true}`), nil
		})}),
	)
	var result bool
	if err := client.Call(context.Background(), "getMe", nil, &result); err != nil {
		t.Fatal(err)
	}
	if !result || path != "/botTOKEN/test/getMe" {
		t.Fatalf("result = %v, path = %q", result, path)
	}
}

func TestClientTestEnvironmentFilePath(t *testing.T) {
	t.Parallel()

	var path string
	client := New("TOKEN",
		WithBaseURL("https://telegram.invalid"),
		WithTestEnvironment(true),
		WithHTTPClient(&http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			path = request.URL.Path
			return testResponse(http.StatusOK, "contents"), nil
		})}),
	)
	file, err := client.OpenFile(context.Background(), "documents/conformance.txt")
	if err != nil {
		t.Fatal(err)
	}
	_ = file.Close()
	if path != "/file/botTOKEN/test/documents/conformance.txt" {
		t.Fatalf("path = %q", path)
	}
}

func TestSuccessfulEnvelopeRequiresResult(t *testing.T) {
	t.Parallel()

	for _, body := range []string{`{"ok":true}`, `{"ok":true,"result":null}`} {
		body := body
		t.Run(body, func(t *testing.T) {
			client := New("TOKEN", WithHTTPClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return testResponse(http.StatusOK, body), nil
			})}))
			err := client.Call(context.Background(), "getMe", nil, nil)
			if !errors.Is(err, ErrResultMissing) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestRemoteErrorsRedactToken(t *testing.T) {
	t.Parallel()

	const token = "SUPER_SECRET_TOKEN"
	tests := []string{
		`not json: SUPER_SECRET_TOKEN`,
		`{"ok":false,"error_code":400,"description":"SUPER_SECRET_TOKEN"}`,
	}
	for _, body := range tests {
		client := New(token, WithHTTPClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return testResponse(http.StatusBadRequest, body), nil
		})}))
		err := client.Call(context.Background(), "getMe", nil, nil)
		if err == nil || strings.Contains(err.Error(), token) {
			t.Fatalf("unsanitized error = %v", err)
		}
	}

	client := New(token, WithHTTPClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("proxy exposed " + token)
	})}))
	err := client.Call(context.Background(), "getMe", nil, nil)
	if err == nil || strings.Contains(err.Error(), token) || !strings.Contains(err.Error(), "<redacted>") {
		t.Fatalf("unsanitized transport error = %v", err)
	}
}

func TestMultipartRejectsHeaderInjection(t *testing.T) {
	t.Parallel()

	client := New("TOKEN")
	tests := []struct {
		name    string
		fields  map[string]string
		uploads []Upload
	}{
		{name: "field", fields: map[string]string{"chat_id\r\nX-Injected": "1"}},
		{name: "upload field", uploads: []Upload{{Field: "photo\r\nX-Injected", Reader: strings.NewReader("x")}}},
		{name: "file name", uploads: []Upload{{Field: "photo", Name: "x\r\nX-Injected", Reader: strings.NewReader("x")}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := client.CallMultipart(context.Background(), "sendPhoto", test.fields, test.uploads, nil); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

type panicReader struct{}

func (panicReader) Read([]byte) (int, error) { panic("reader panic") }

func TestMultipartContainsReaderPanic(t *testing.T) {
	t.Parallel()

	client := New("TOKEN", WithHTTPClient(&http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		_, err := io.Copy(io.Discard, request.Body)
		if err != nil {
			return nil, err
		}
		return testResponse(http.StatusOK, `{"ok":true,"result":true}`), nil
	})}))
	var result bool
	err := client.CallMultipart(context.Background(), "sendPhoto", nil, []Upload{{
		Field: "photo", Name: "photo.jpg", Reader: panicReader{},
	}}, &result)
	if err == nil || !strings.Contains(err.Error(), "multipart writer panicked") {
		t.Fatalf("error = %v", err)
	}
}

func testResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
