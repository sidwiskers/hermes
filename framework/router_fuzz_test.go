package framework

import (
	"context"
	"strings"
	"testing"

	telegram "github.com/sidwiskers/hermes/types"
)

func FuzzCallbackPrefixIndex(f *testing.F) {
	f.Add("item:999:value", "item:", "item:999:", "other:", uint8(0b111))
	f.Add("same:value", "same:", "same:", "s", uint8(0b110))
	f.Add("", "a", "ab", "abc", uint8(0b111))

	f.Fuzz(func(t *testing.T, data, first, second, third string, enabled uint8) {
		if len(data) > 256 {
			data = data[:256]
		}
		prefixes := []string{first, second, third}
		routes := make([]prefixRouteDef, 0, len(prefixes))
		selected := -1
		for index, prefix := range prefixes {
			if len(prefix) > 256 {
				prefix = prefix[:256]
				prefixes[index] = prefix
			}
			if prefix == "" {
				continue
			}
			routeIndex := len(routes)
			allowed := enabled&(1<<index) != 0
			routes = append(routes, prefixRouteDef{
				prefix: prefix,
				route: routeDef{
					filter: func(*Context) bool { return allowed },
					handler: func(*Context) error {
						selected = routeIndex
						return nil
					},
				},
			})
		}

		table := compileCallbackPrefixes(routes, nil)
		ctx := NewContext(context.Background(), nil, &telegram.Update{
			CallbackQuery: &telegram.CallbackQuery{Data: data},
		}, "")
		if handler := matchCallbackPrefix(table, data, ctx); handler != nil {
			if err := handler(ctx); err != nil {
				t.Fatal(err)
			}
		}

		expected := -1
		longest := -1
		routeIndex := 0
		for index, prefix := range prefixes {
			if prefix == "" {
				continue
			}
			if enabled&(1<<index) != 0 && strings.HasPrefix(data, prefix) && len(prefix) > longest {
				expected = routeIndex
				longest = len(prefix)
			}
			routeIndex++
		}
		if selected != expected {
			t.Fatalf(
				"data=%q prefixes=%q enabled=%03b selected=%d expected=%d",
				data, prefixes, enabled, selected, expected,
			)
		}
	})
}
