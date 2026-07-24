package api

import (
	"net/url"
	"path"
	"strings"
	"testing"
)

func FuzzCleanTelegramFilePath(f *testing.F) {
	f.Add("photos/a.jpg")
	f.Add("../token")
	f.Add("%2e%2e/token")
	f.Add("documents/%5csecret")

	f.Fuzz(func(t *testing.T, value string) {
		clean, err := cleanTelegramFilePath(value)
		if err != nil {
			return
		}
		decoded, err := url.PathUnescape(clean)
		if err != nil {
			t.Fatalf("accepted path cannot be decoded: %q", clean)
		}
		normalized := path.Clean(decoded)
		if clean == "" ||
			strings.ContainsAny(clean, "\\?#") ||
			strings.HasPrefix(decoded, "/") ||
			strings.Contains(decoded, "\\") ||
			normalized == "." ||
			normalized == ".." ||
			strings.HasPrefix(normalized, "../") {
			t.Fatalf("unsafe path accepted: input=%q clean=%q decoded=%q", value, clean, decoded)
		}
		for _, candidate := range []string{clean, decoded} {
			for index := 0; index < len(candidate); index++ {
				if candidate[index] < 0x20 || candidate[index] == 0x7f {
					t.Fatalf("control character accepted in %q", candidate)
				}
			}
		}
	})
}
