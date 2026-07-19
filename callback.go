package hermes

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// MaxCallbackDataBytes is Telegram's UTF-8 byte limit for callback data.
const MaxCallbackDataBytes = 64

var (
	// ErrCallbackPrefixMismatch means callback data did not begin with the
	// codec's configured prefix.
	ErrCallbackPrefixMismatch = errors.New("hermes: callback prefix mismatch")
	// ErrCallbackDataTooLong means encoded callback data exceeded Telegram's
	// 64-byte limit.
	ErrCallbackDataTooLong = errors.New("hermes: callback data exceeds 64 bytes")
)

// CallbackCodec keeps callback-data formatting next to its parser.
// Prefix should include a separator, for example "user:".
type CallbackCodec[T any] struct {
	Prefix string
	Encode func(T) (string, error)
	Decode func(string) (T, error)
}

// Data encodes value, prepends Prefix, and enforces Telegram's size limit.
func (codec CallbackCodec[T]) Data(value T) (string, error) {
	var zero string
	if codec.Encode == nil {
		return zero, fmt.Errorf("hermes: callback codec has no encoder")
	}
	payload, err := codec.Encode(value)
	if err != nil {
		return zero, err
	}
	data := codec.Prefix + payload
	if len(data) > MaxCallbackDataBytes {
		return zero, ErrCallbackDataTooLong
	}
	return data, nil
}

// MustData is Data for static configuration. It panics if encoding fails.
func (codec CallbackCodec[T]) MustData(value T) string {
	data, err := codec.Data(value)
	if err != nil {
		panic(err)
	}
	return data
}

// Button creates a callback button containing the encoded value.
func (codec CallbackCodec[T]) Button(text string, value T) (InlineKeyboardButton, error) {
	data, err := codec.Data(value)
	if err != nil {
		return InlineKeyboardButton{}, err
	}
	return Button(text, data), nil
}

// MustButton is Button for static configuration. It panics if encoding fails.
func (codec CallbackCodec[T]) MustButton(text string, value T) InlineKeyboardButton {
	button, err := codec.Button(text, value)
	if err != nil {
		panic(err)
	}
	return button
}

// Parse validates Prefix and decodes a callback-data string.
func (codec CallbackCodec[T]) Parse(data string) (T, error) {
	var zero T
	if !strings.HasPrefix(data, codec.Prefix) {
		return zero, ErrCallbackPrefixMismatch
	}
	if codec.Decode == nil {
		return zero, fmt.Errorf("hermes: callback codec has no decoder")
	}
	return codec.Decode(strings.TrimPrefix(data, codec.Prefix))
}

// Handler adapts a typed callback handler to Handler.
func (codec CallbackCodec[T]) Handler(handler func(*Context, T) error) Handler {
	return func(c *Context) error {
		if handler == nil {
			return fmt.Errorf("hermes: callback codec handler is nil")
		}
		if c == nil || c.Callback == nil {
			return ErrCallbackPrefixMismatch
		}
		value, err := codec.Parse(c.Callback.Data)
		if err != nil {
			return err
		}
		return handler(c, value)
	}
}

// StringCallback returns a codec whose payload is stored verbatim.
func StringCallback(prefix string) CallbackCodec[string] {
	return CallbackCodec[string]{
		Prefix: prefix,
		Encode: func(value string) (string, error) { return value, nil },
		Decode: func(value string) (string, error) { return value, nil },
	}
}

// IntCallback returns a base-10 int callback codec.
func IntCallback(prefix string) CallbackCodec[int] {
	return CallbackCodec[int]{
		Prefix: prefix,
		Encode: func(value int) (string, error) { return strconv.Itoa(value), nil },
		Decode: strconv.Atoi,
	}
}

// Int64Callback returns a base-10 int64 callback codec.
func Int64Callback(prefix string) CallbackCodec[int64] {
	return CallbackCodec[int64]{
		Prefix: prefix,
		Encode: func(value int64) (string, error) { return strconv.FormatInt(value, 10), nil },
		Decode: func(value string) (int64, error) { return strconv.ParseInt(value, 10, 64) },
	}
}

// JSONCallback is convenient for compact structs. Telegram limits callback
// data to 64 bytes, and Data enforces that limit before the button is sent.
func JSONCallback[T any](prefix string) CallbackCodec[T] {
	return CallbackCodec[T]{
		Prefix: prefix,
		Encode: func(value T) (string, error) {
			data, err := json.Marshal(value)
			return string(data), err
		},
		Decode: func(value string) (T, error) {
			var decoded T
			err := json.Unmarshal([]byte(value), &decoded)
			return decoded, err
		},
	}
}
