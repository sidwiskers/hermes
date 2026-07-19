package api

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type messageOrBool struct {
	Message *Message
	OK      bool
}

func (r *messageOrBool) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 4 && data[0] == 't' && data[1] == 'r' && data[2] == 'u' && data[3] == 'e' {
		r.OK = true
		return nil
	}
	if len(data) == 5 && data[0] == 'f' && data[1] == 'a' && data[2] == 'l' && data[3] == 's' && data[4] == 'e' {
		return nil
	}

	var message Message
	if err := json.Unmarshal(data, &message); err != nil {
		return fmt.Errorf("hermes: expected Message or boolean: %w", err)
	}
	r.Message = &message
	r.OK = true
	return nil
}
