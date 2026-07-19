package hermes

import (
	"errors"
	"testing"
)

func TestTypedCallbackCodec(t *testing.T) {
	t.Parallel()

	codec := Int64Callback("user:")
	data, err := codec.Data(922337)
	if err != nil {
		t.Fatal(err)
	}
	if data != "user:922337" {
		t.Fatalf("data = %q", data)
	}
	value, err := codec.Parse(data)
	if err != nil || value != 922337 {
		t.Fatalf("parse = %d, %v", value, err)
	}
}

func TestCallbackCodecEnforcesTelegramLimit(t *testing.T) {
	t.Parallel()

	codec := StringCallback("x:")
	_, err := codec.Data(string(make([]byte, MaxCallbackDataBytes)))
	if !errors.Is(err, ErrCallbackDataTooLong) {
		t.Fatalf("expected size error, got %v", err)
	}
}
