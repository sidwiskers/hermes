package types

import (
	"bytes"
	"encoding/json"
	"testing"
)

var updateFixture = []byte(`{"update_id":42,"message":{"message_id":7,"date":1,"chat":{"id":99,"type":"private"},"from":{"id":5,"is_bot":false,"first_name":"Ada"},"text":"/start hello"},"future_field":{"enabled":true}}`)

func TestDecodeUpdateModes(t *testing.T) {
	t.Parallel()

	fast, err := DecodeUpdate(updateFixture, false)
	if err != nil {
		t.Fatal(err)
	}
	if fast.UpdateID != 42 || fast.Message == nil || len(fast.Raw) != 0 {
		t.Fatalf("fast update = %#v", fast)
	}

	preserved, err := DecodeUpdate(updateFixture, true)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(preserved.Raw, updateFixture) {
		t.Fatalf("raw = %s", preserved.Raw)
	}
}

func TestDecodeUpdateDoesNotAliasPooledEnvelope(t *testing.T) {
	first, err := DecodeUpdate(updateFixture, false)
	if err != nil {
		t.Fatal(err)
	}
	second, err := DecodeUpdate([]byte(`{"update_id":2}`), false)
	if err != nil {
		t.Fatal(err)
	}
	if first.UpdateID == second.UpdateID || first.Message == nil || first.Message.Text != "/start hello" {
		t.Fatalf("first decode changed after pool reuse: %#v", first)
	}
}

func TestDecodeUpdatesPreservesEachElement(t *testing.T) {
	t.Parallel()

	payload := []byte(`[{"update_id":1,"x":1},{"update_id":2,"x":2}]`)
	updates, err := DecodeUpdates(payload, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(updates) != 2 || updates[0].UpdateID != 1 || updates[1].UpdateID != 2 {
		t.Fatalf("updates = %#v", updates)
	}
	if string(updates[0].Raw) != `{"update_id":1,"x":1}` || string(updates[1].Raw) != `{"update_id":2,"x":2}` {
		t.Fatalf("raw elements = %q, %q", updates[0].Raw, updates[1].Raw)
	}
}

func TestDefaultJSONDecodeDoesNotRetainRaw(t *testing.T) {
	t.Parallel()

	var update Update
	if err := json.Unmarshal(updateFixture, &update); err != nil {
		t.Fatal(err)
	}
	if len(update.Raw) != 0 {
		t.Fatalf("default decode retained %d raw bytes", len(update.Raw))
	}
}

func BenchmarkDecodeUpdateFast(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		update, err := DecodeUpdate(updateFixture, false)
		if err != nil || update.UpdateID != 42 {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeUpdateStdlib(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var update Update
		if err := json.Unmarshal(updateFixture, &update); err != nil || update.UpdateID != 42 {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeUpdatePreserveRaw(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		update, err := DecodeUpdate(updateFixture, true)
		if err != nil || len(update.Raw) == 0 {
			b.Fatal(err)
		}
	}
}

func FuzzDecodeUpdate(f *testing.F) {
	f.Add(updateFixture, false)
	f.Add(updateFixture, true)
	f.Add([]byte(`{"update_id":1}`), false)
	f.Fuzz(func(t *testing.T, data []byte, preserve bool) {
		update, err := DecodeUpdate(data, preserve)
		if err != nil {
			return
		}
		if preserve && !bytes.Equal(update.Raw, data) {
			t.Fatalf("preserved raw mismatch")
		}
		if !preserve && len(update.Raw) != 0 {
			t.Fatalf("fast mode retained raw")
		}
	})
}
