package botapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var (
	headingPattern = regexp.MustCompile(`(?s)<h4><a class="anchor" name="([^"]+)"[^>]*>.*?</a>(.*?)</h4>`)
	tablePattern   = regexp.MustCompile(`(?s)<table class="table">(.*?)</table>`)
	rowPattern     = regexp.MustCompile(`(?s)<tr>(.*?)</tr>`)
	cellPattern    = regexp.MustCompile(`(?s)<t[dh][^>]*>(.*?)</t[dh]>`)
	tagPattern     = regexp.MustCompile(`(?s)<[^>]+>`)
	listPattern    = regexp.MustCompile(`(?s)<li[^>]*>(.*?)</li>`)
	linkPattern    = regexp.MustCompile(`(?s)<a[^>]+href="#([^"]+)"[^>]*>(.*?)</a>`)
	versionPattern = regexp.MustCompile(`Bot API ([0-9]+\.[0-9]+)`)
)

type heading struct {
	anchor  string
	title   string
	section []byte
}

func ParseHTML(data []byte) (Manifest, error) {
	headings := extractHeadings(data)
	if len(headings) == 0 {
		return Manifest{}, fmt.Errorf("botapi: no Bot API headings found")
	}
	titles := make(map[string]string, len(headings))
	for _, item := range headings {
		titles[item.anchor] = item.title
	}

	digest := sha256.Sum256(data)
	result := Manifest{
		Released:     releaseDate(headings),
		Source:       OfficialSource,
		SourceSHA256: hex.EncodeToString(digest[:]),
	}
	if match := versionPattern.FindSubmatch(data); len(match) == 2 {
		result.Version = string(match[1])
	}
	for _, item := range headings {
		headers, rows := extractTable(item.section)
		switch {
		case isMethodName(item.title):
			parameters := []Field{}
			if equalStrings(headers, []string{"Parameter", "Type", "Required", "Description"}) {
				for _, row := range rows {
					if len(row) != 4 {
						return Manifest{}, fmt.Errorf("botapi: method %s has malformed row", item.title)
					}
					parameters = append(parameters, Field{Name: row[0], Type: row[1], Required: row[2] == "Yes"})
				}
			} else if len(headers) != 0 {
				return Manifest{}, fmt.Errorf("botapi: method %s has unexpected table headers %q", item.title, headers)
			}
			result.Methods = append(result.Methods, Method{Name: item.title, Anchor: item.anchor, Parameters: parameters})
		case equalStrings(headers, []string{"Field", "Type", "Description"}):
			fields := make([]Field, 0, len(rows))
			for _, row := range rows {
				if len(row) != 3 {
					return Manifest{}, fmt.Errorf("botapi: object %s has malformed row", item.title)
				}
				fields = append(fields, Field{Name: row[0], Type: row[1], Required: !strings.HasPrefix(row[2], "Optional")})
			}
			result.Objects = append(result.Objects, Object{Name: item.title, Anchor: item.anchor, Fields: fields})
		case len(headers) == 0:
			variants := unionVariants(item.section, titles)
			if len(variants) != 0 {
				entry := Union{Name: item.title, Anchor: item.anchor, Variants: variants}
				if item.title == "RichText" {
					entry.Alternatives = []string{"String", "Array of RichText"}
				}
				result.Unions = append(result.Unions, entry)
			} else if isEmptyOrAbstractObject(item.title, plainText(item.section)) {
				result.Objects = append(result.Objects, Object{Name: item.title, Anchor: item.anchor, Fields: []Field{}})
			}
		}
	}

	sort.Slice(result.Methods, func(i, j int) bool { return result.Methods[i].Name < result.Methods[j].Name })
	sort.Slice(result.Objects, func(i, j int) bool { return result.Objects[i].Name < result.Objects[j].Name })
	sort.Slice(result.Unions, func(i, j int) bool { return result.Unions[i].Name < result.Unions[j].Name })
	if result.Version == "" || len(result.Methods) == 0 || len(result.Objects) == 0 {
		return Manifest{}, fmt.Errorf("botapi: incomplete schema: version=%q methods=%d objects=%d", result.Version, len(result.Methods), len(result.Objects))
	}
	return result, nil
}

func extractHeadings(data []byte) []heading {
	matches := headingPattern.FindAllSubmatchIndex(data, -1)
	result := make([]heading, 0, len(matches))
	for index, match := range matches {
		sectionEnd := len(data)
		if index+1 < len(matches) {
			sectionEnd = matches[index+1][0]
		}
		result = append(result, heading{
			anchor:  string(data[match[2]:match[3]]),
			title:   plainText(data[match[4]:match[5]]),
			section: data[match[1]:sectionEnd],
		})
	}
	return result
}

func extractTable(section []byte) ([]string, [][]string) {
	table := tablePattern.FindSubmatch(section)
	if len(table) != 2 {
		return nil, nil
	}
	rows := rowPattern.FindAllSubmatch(table[1], -1)
	parsed := make([][]string, 0, len(rows))
	for _, row := range rows {
		cells := cellPattern.FindAllSubmatch(row[1], -1)
		values := make([]string, 0, len(cells))
		for _, cell := range cells {
			values = append(values, plainText(cell[1]))
		}
		if len(values) != 0 {
			parsed = append(parsed, values)
		}
	}
	if len(parsed) == 0 {
		return nil, nil
	}
	return parsed[0], parsed[1:]
}

func unionVariants(section []byte, titles map[string]string) []string {
	items := listPattern.FindAllSubmatch(section, -1)
	variants := make([]string, 0, len(items))
	for _, item := range items {
		links := linkPattern.FindAllSubmatch(item[1], -1)
		if len(links) != 1 {
			return nil
		}
		anchor := string(links[0][1])
		title, ok := titles[anchor]
		if !ok || plainText(links[0][2]) != title || plainText(item[1]) != title {
			return nil
		}
		variants = append(variants, title)
	}
	return variants
}

func isMethodName(name string) bool {
	for _, r := range name {
		return unicode.IsLower(r)
	}
	return false
}

func isEmptyOrAbstractObject(name, description string) bool {
	if name == "InputFile" || name == "CallbackGame" {
		return true
	}
	return strings.Contains(description, "holds no information")
}

func releaseDate(headings []heading) string {
	for _, item := range headings {
		if versionPattern.Match(item.section) {
			return item.title
		}
	}
	return ""
}

func plainText(data []byte) string {
	withoutTags := tagPattern.ReplaceAll(data, []byte(" "))
	return strings.Join(strings.Fields(html.UnescapeString(string(bytes.TrimSpace(withoutTags)))), " ")
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
