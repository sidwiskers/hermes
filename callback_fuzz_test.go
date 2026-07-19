package hermes

import "testing"

func FuzzInt64CallbackRoundTrip(f *testing.F) {
	f.Add("item:", int64(42))
	f.Add("", int64(-1))
	f.Fuzz(func(t *testing.T, prefix string, value int64) {
		codec := Int64Callback(prefix)
		data, err := codec.Data(value)
		if err != nil {
			return
		}
		decoded, err := codec.Parse(data)
		if err != nil || decoded != value {
			t.Fatalf("round trip: got=%d want=%d err=%v", decoded, value, err)
		}
	})
}
