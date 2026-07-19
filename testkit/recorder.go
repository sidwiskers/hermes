// Package testkit provides a deterministic in-memory Bot API transport for tests.
package testkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"

	"github.com/sidwiskers/hermes"
	"github.com/sidwiskers/hermes/api"
)

// File is one uploaded multipart file captured by Recorder.
type File struct {
	Name string
	Data []byte
}

// Request is a decoded Bot API request captured by Recorder.
type Request struct {
	Method  string
	Header  http.Header
	JSON    map[string]any
	Form    map[string]string
	Files   map[string]File
	RawBody []byte
}

// Response describes one queued Telegram-style response.
type Response struct {
	StatusCode  int
	Result      any
	ErrorCode   int
	Description string
	Parameters  *hermes.ResponseParameters
}

// Recorder implements http.RoundTripper and records requests in arrival order.
type Recorder struct {
	mu        sync.Mutex
	requests  []Request
	responses []Response
}

// New creates a framework bot and its in-memory request recorder.
func New() (*hermes.Bot, *Recorder) {
	recorder := &Recorder{}
	httpClient := &http.Client{Transport: recorder}
	bot := hermes.New("TEST_TOKEN", hermes.WithBaseURL("https://telegram.invalid"), hermes.WithHTTPClient(httpClient))
	return bot, recorder
}

// NewClient creates the standalone low-level API client with the same recorder.
func NewClient() (*api.Client, *Recorder) {
	recorder := &Recorder{}
	httpClient := &http.Client{Transport: recorder}
	client := api.New("TEST_TOKEN", api.WithBaseURL("https://telegram.invalid"), api.WithHTTPClient(httpClient))
	return client, recorder
}

// Respond queues a successful response containing result.
func (r *Recorder) Respond(result any) {
	r.Enqueue(Response{StatusCode: http.StatusOK, Result: result})
}

// Fail queues a Telegram API error response.
func (r *Recorder) Fail(code int, description string, parameters *hermes.ResponseParameters) {
	r.Enqueue(Response{StatusCode: code, ErrorCode: code, Description: description, Parameters: parameters})
}

// Enqueue appends a custom response. Responses are consumed in FIFO order.
func (r *Recorder) Enqueue(response Response) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.responses = append(r.responses, response)
}

// Requests returns a snapshot of every request received so far.
func (r *Recorder) Requests() []Request {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]Request, len(r.requests))
	copy(result, r.requests)
	return result
}

// Last returns the most recently captured request.
func (r *Recorder) Last() (Request, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.requests) == 0 {
		return Request{}, false
	}
	return r.requests[len(r.requests)-1], true
}

// Reset removes all captured requests and queued responses.
func (r *Recorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = nil
	r.responses = nil
}

// RoundTrip implements http.RoundTripper.
func (r *Recorder) RoundTrip(request *http.Request) (*http.Response, error) {
	recorded, err := decodeRequest(request)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.requests = append(r.requests, recorded)
	response := Response{StatusCode: http.StatusOK, Result: true}
	if len(r.responses) != 0 {
		response = r.responses[0]
		r.responses = r.responses[1:]
	}
	r.mu.Unlock()

	status := response.StatusCode
	if status == 0 {
		status = http.StatusOK
	}
	envelope := map[string]any{}
	if response.ErrorCode != 0 || status >= 400 {
		envelope["ok"] = false
		envelope["error_code"] = response.ErrorCode
		if response.ErrorCode == 0 {
			envelope["error_code"] = status
		}
		envelope["description"] = response.Description
		if response.Parameters != nil {
			envelope["parameters"] = response.Parameters
		}
	} else {
		envelope["ok"] = true
		envelope["result"] = response.Result
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(body)),
		Request:    request,
	}, nil
}

func decodeRequest(request *http.Request) (Request, error) {
	method := strings.TrimPrefix(request.URL.Path, "/")
	if index := strings.LastIndex(request.URL.Path, "/"); index >= 0 {
		method = request.URL.Path[index+1:]
	}
	result := Request{
		Method: method,
		Header: request.Header.Clone(),
		Form:   make(map[string]string),
		Files:  make(map[string]File),
	}
	if request.Body == nil {
		return result, nil
	}
	contentType, parameters, _ := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if contentType == "multipart/form-data" {
		reader := multipart.NewReader(request.Body, parameters["boundary"])
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return Request{}, err
			}
			data, err := io.ReadAll(part)
			part.Close()
			if err != nil {
				return Request{}, err
			}
			if part.FileName() == "" {
				result.Form[part.FormName()] = string(data)
			} else {
				result.Files[part.FormName()] = File{Name: part.FileName(), Data: data}
			}
		}
		return result, nil
	}
	data, err := io.ReadAll(request.Body)
	if err != nil {
		return Request{}, err
	}
	result.RawBody = data
	if len(bytes.TrimSpace(data)) != 0 {
		if err := json.Unmarshal(data, &result.JSON); err != nil {
			return Request{}, err
		}
	}
	return result, nil
}
