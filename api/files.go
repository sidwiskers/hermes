package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

type GetFileParams struct {
	FileID string `json:"file_id"`
}

func (b *Client) GetFile(ctx context.Context, fileID string) (*File, error) {
	if strings.TrimSpace(fileID) == "" {
		return nil, fmt.Errorf("hermes: getFile file_id is required")
	}
	var file File
	if err := b.Call(ctx, "getFile", GetFileParams{FileID: fileID}, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

// OpenFile opens a Telegram file download as a stream. The caller must close it.
func (b *Client) OpenFile(ctx context.Context, filePath string) (io.ReadCloser, error) {
	if b == nil {
		return nil, ErrClientRequired
	}
	if b.token == "" {
		return nil, ErrTokenRequired
	}
	clean, err := cleanTelegramFilePath(filePath)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, b.fileURL(clean), nil)
	if err != nil {
		return nil, b.transportError("getFile", "create download request", err)
	}
	request.Header.Set("User-Agent", b.userAgent)
	var (
		event           CallEvent
		started         time.Time
		observedContext = request.Context()
	)
	if b.observer != nil {
		event = CallEvent{Method: "downloadFile", Kind: CallDownload}
		started = time.Now()
		observedContext = startObserver(b.observer, observedContext, event)
		request = request.WithContext(observedContext)
	}

	response, err := b.client.Do(request)
	if err != nil {
		resultErr := b.transportError("getFile", "download", err)
		if b.observer != nil {
			finishObserver(b.observer, observedContext, event, CallResult{Duration: time.Since(started), Err: resultErr})
		}
		return nil, resultErr
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		defer response.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(response.Body, 513))
		resultErr := &HTTPError{
			StatusCode: response.StatusCode,
			Status:     redactToken(response.Status, b.token),
			Body:       b.compactBody(body),
		}
		if b.observer != nil {
			finishObserver(b.observer, observedContext, event, CallResult{Duration: time.Since(started), Err: resultErr})
		}
		return nil, resultErr
	}
	if b.observer == nil {
		return response.Body, nil
	}
	return &observedDownload{
		ReadCloser: response.Body,
		finish: func(err error) {
			finishObserver(b.observer, observedContext, event, CallResult{
				Duration: time.Since(started),
				Err:      err,
			})
		},
	}, nil
}

func (b *Client) DownloadFile(ctx context.Context, filePath string, destination io.Writer) (int64, error) {
	if b == nil {
		return 0, ErrClientRequired
	}
	if destination == nil {
		return 0, fmt.Errorf("hermes: download destination is required")
	}
	reader, err := b.OpenFile(ctx, filePath)
	if err != nil {
		return 0, err
	}
	defer reader.Close()
	written, err := io.Copy(destination, reader)
	if err != nil {
		return written, b.transportError("getFile", "copy download", err)
	}
	return written, nil
}

func (b *Client) fileURL(filePath string) string {
	return b.filePrefix + filePath
}

func cleanTelegramFilePath(value string) (string, error) {
	value = strings.TrimSpace(strings.TrimPrefix(value, "/"))
	if value == "" || strings.Contains(value, "\\") {
		return "", fmt.Errorf("hermes: invalid Telegram file path")
	}
	clean := path.Clean(value)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("hermes: invalid Telegram file path")
	}
	return clean, nil
}

type observedDownload struct {
	io.ReadCloser
	once   sync.Once
	finish func(error)
}

func (r *observedDownload) Read(buffer []byte) (int, error) {
	read, err := r.ReadCloser.Read(buffer)
	if err != nil {
		if err == io.EOF {
			r.complete(nil)
		} else {
			r.complete(err)
		}
	}
	return read, err
}

func (r *observedDownload) Close() error {
	err := r.ReadCloser.Close()
	r.complete(err)
	return err
}

func (r *observedDownload) complete(err error) {
	r.once.Do(func() {
		if r.finish != nil {
			r.finish(err)
		}
	})
}
