package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
)

// json.Unmarshal must receive a pointer, which otherwise moves the relatively
// large Update envelope to the heap on every call. Nested payloads remain
// independently owned after the envelope is copied out of the pool.
var updateDecodePool = sync.Pool{New: func() any { return new(Update) }}

// DecodeUpdate decodes one Telegram update. preserveRaw controls whether the
// complete original JSON payload is copied into Update.Raw.
func DecodeUpdate(data []byte, preserveRaw bool) (Update, error) {
	pooled := updateDecodePool.Get().(*Update)
	*pooled = Update{}
	if err := json.Unmarshal(data, pooled); err != nil {
		*pooled = Update{}
		updateDecodePool.Put(pooled)
		return Update{}, fmt.Errorf("hermes: decode update: %w", err)
	}
	update := *pooled
	*pooled = Update{}
	updateDecodePool.Put(pooled)
	if preserveRaw {
		update.Raw = bytes.Clone(data)
	}
	return update, nil
}

// DecodeUpdates decodes a getUpdates result. In fast mode it performs one
// direct slice decode. Raw-preserving mode intentionally decodes each element
// independently so each Update.Raw contains exactly that update's JSON.
func DecodeUpdates(data []byte, preserveRaw bool) ([]Update, error) {
	if !preserveRaw {
		var updates []Update
		if err := json.Unmarshal(data, &updates); err != nil {
			return nil, fmt.Errorf("hermes: decode updates: %w", err)
		}
		return updates, nil
	}

	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("hermes: decode updates: %w", err)
	}
	updates := make([]Update, len(raw))
	for index := range raw {
		update, err := DecodeUpdate(raw[index], true)
		if err != nil {
			return nil, fmt.Errorf("hermes: decode update %d: %w", index, err)
		}
		updates[index] = update
	}
	return updates, nil
}
