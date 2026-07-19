package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

var attachmentJSONPrefix = []byte(`"attach://`)

func validateAttachmentUploads(value any, uploads []Upload, method string) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("hermes: encode %s attachments: %w", method, err)
	}
	references := make(map[string]struct{})
	if err := collectAttachmentReferencesJSON(data, references); err != nil {
		return fmt.Errorf("hermes: inspect %s attachments: %w", method, err)
	}

	provided := make(map[string]struct{}, len(uploads))
	for _, upload := range uploads {
		field := strings.TrimSpace(upload.Field)
		if field == "" || upload.Reader == nil {
			return fmt.Errorf("hermes: invalid %s upload", method)
		}
		if _, exists := provided[field]; exists {
			return fmt.Errorf("hermes: duplicate multipart field %q", field)
		}
		provided[field] = struct{}{}
	}

	for field := range references {
		if _, exists := provided[field]; !exists {
			return fmt.Errorf("hermes: %s attachment %q has no upload", method, field)
		}
	}
	for field := range provided {
		if _, exists := references[field]; !exists {
			return fmt.Errorf("hermes: %s upload %q is not referenced", method, field)
		}
	}
	return nil
}

func collectAttachmentReferencesJSON(data []byte, fields map[string]struct{}) error {
	for offset := 0; offset < len(data); {
		relative := bytes.Index(data[offset:], attachmentJSONPrefix)
		if relative < 0 {
			return nil
		}
		quote := offset + relative
		if jsonQuoteEscaped(data, quote) {
			offset = quote + 1
			continue
		}

		start := quote + len(attachmentJSONPrefix)
		end := start
		for end < len(data) {
			switch data[end] {
			case '\\':
				end += 2
				continue
			case '"':
				break
			}
			if data[end] == '"' {
				break
			}
			end++
		}
		if end >= len(data) {
			return strconv.ErrSyntax
		}
		var field string
		if bytes.IndexByte(data[start:end], '\\') >= 0 {
			var value string
			if err := json.Unmarshal(data[quote:end+1], &value); err != nil {
				return err
			}
			field = strings.TrimSpace(strings.TrimPrefix(value, "attach://"))
		} else {
			field = strings.TrimSpace(string(data[start:end]))
		}
		if field != "" {
			fields[field] = struct{}{}
		}
		offset = end + 1
	}
	return nil
}

func jsonQuoteEscaped(data []byte, quote int) bool {
	backslashes := 0
	for index := quote - 1; index >= 0 && data[index] == '\\'; index-- {
		backslashes++
	}
	return backslashes%2 != 0
}
